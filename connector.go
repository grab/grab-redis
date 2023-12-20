// MIT License
//
//
// Copyright 2023 Grabtaxi Holdings Pte Ltd (GRAB), All rights reserved.
//
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE

package redis

import (
	"context"
	"fmt"
	"strings"

	"github.com/myteksi/hystrix-go/hystrix"
	"github.com/pkg/errors"

	"github.com/grab/grab-redis/circuitbreaker"
	"github.com/grab/grab-redis/redisapi"
)

type connectorImpl struct {
	client          *clientImpl
	loadTestClients []*clientImpl

	processAllLoadTestPackets bool
	schedulerOptions          *schedulerOptions
	loadTestScheduler         *scheduler
	schedulerCancel           context.CancelFunc

	configurer Configurer
	stats      StatsClient
	logger     Logger
	cbOptions  []circuitbreaker.Option
}

func NewStaticConnector(ctx context.Context, config *ConnectorConfig, options ...ConnectorOption) (redisapi.Client, error) {
	return NewDynamicConnector(ctx, newStaticConfigurer(config), options...)
}

func NewDynamicConnector(ctx context.Context, configurer Configurer, options ...ConnectorOption) (redisapi.Client, error) {
	c := &connectorImpl{
		schedulerOptions: &schedulerOptions{
			maxChanSize:       defaultMaxChanSize,
			maxWorker:         defaultMaxWorker,
			workerIdleTimeout: defaultWorkerIdleTimeout,
		},

		configurer: configurer,
		cbOptions:  getDefaultCBOptions(),
	}

	// apply options to the client
	for _, opt := range options {
		opt(c)
	}

	var err error
	config := &ConnectorConfig{}
	if err = c.configurer.Unmarshal(config); err != nil {
		return nil, err
	}
	if err = config.initAndValidate(); err != nil {
		return nil, err
	}

	c.client, err = newClient(ctx, config.Main, ClientStatsD(c.stats), ClientLogger(c.logger), ClientCBOptions(c.cbOptions))
	if err != nil {
		return nil, err
	}

	c.loadTestClients = make([]*clientImpl, len(config.LoadTests))
	for i, config := range config.LoadTests {
		c.loadTestClients[i], err = newClient(ctx, config, ClientStatsD(c.stats), ClientLogger(c.logger), ClientCBOptions(c.cbOptions))
		if err != nil {
			return nil, err
		}
	}
	// The callback function which is used to reload the configuration dynamically
	c.configurer.OnChange(func() error {
		connectorOptions := &ConnectorConfig{}
		if err = c.configurer.Unmarshal(connectorOptions); err != nil {
			return err
		}
		if err = connectorOptions.initAndValidate(); err != nil {
			return err
		}
		return c.reload(ctx, connectorOptions)
	})

	c.processAllLoadTestPackets = config.ProcessAllLoadTestPackets
	c.schedulerOptions.maxChanSize = config.SchedulerChannelSize
	c.schedulerOptions.maxWorker = config.SchedulerWorkerNumber
	c.schedulerOptions.workerIdleTimeout = parseDurationInMs(config.SchedulerWorkerIdleTimeoutInMs)
	c.loadTestScheduler = newScheduler(c.schedulerOptions)

	schedulerCtx, cancel := context.WithCancel(ctx)
	go c.loadTestScheduler.start(schedulerCtx)
	c.schedulerCancel = cancel

	return c, nil
}

func (c *connectorImpl) reload(ctx context.Context, config *ConnectorConfig) error {
	if !config.HotReload {
		return nil
	}

	c.processAllLoadTestPackets = config.ProcessAllLoadTestPackets
	if c.schedulerOptions.maxWorker != config.SchedulerWorkerNumber || c.schedulerOptions.maxChanSize != config.SchedulerChannelSize || c.schedulerOptions.workerIdleTimeout != parseDurationInMs(config.SchedulerWorkerIdleTimeoutInMs) {
		return fmt.Errorf("dual write worker number/channel size change is not allowed in reloading")
	}

	var err error

	// TODO: send config to Doorman

	if err = c.client.reload(config.Main); err != nil {
		c.logger.Warn(pkgName, "unable to reload client, using back old client, Error: %s", err)
		return err
	}

	loadTestMap := make(map[string][]*clientImpl)
	for _, client := range c.loadTestClients {
		loadTestMap[client.config.name()] = append(loadTestMap[client.config.name()], client)
	}

	newLoadTestClients := make([]*clientImpl, len(config.LoadTests))
	for i, config := range config.LoadTests {
		var client *clientImpl

		clients := loadTestMap[config.name()]
		if len(clients) > 0 {
			client, loadTestMap[config.name()] = clients[len(clients)-1], clients[:len(clients)-1]
			if isAddrsEquals(c.client.config.Addrs, config.Addrs) {
				c.logger.Warn(pkgName, "unable to reload load test client, fallback to old client, Error: %s", err)
				return fmt.Errorf("can't share the same address with the main client")
			}
			if err = client.reload(config); err != nil {
				c.logger.Warn(pkgName, "unable to reload load test client, fallback to old client, Error: %s", err)
				return err
			}
		} else {
			client, err = newClient(ctx, config, ClientStatsD(c.stats), ClientLogger(c.logger), ClientCBOptions(c.cbOptions))
			if err != nil {
				c.logger.Warn(pkgName, "unable to create new load test client, Error: %s", err)
				return err
			}
		}

		newLoadTestClients[i] = client
	}

	for _, clients := range loadTestMap {
		for _, client := range clients {
			client.ShutDown(ctx)
		}
	}
	c.loadTestClients = newLoadTestClients

	return nil
}

func (c *connectorImpl) queueLoadTest(fn func(context.Context, *clientImpl)) {
	for _, client := range c.loadTestClients {
		if c.processAllLoadTestPackets {
			c.loadTestScheduler.fnChan <- func(ctx context.Context) {
				select {
				case <-ctx.Done(): // This case is executed if ctx is cancelled
					c.logger.Error(pkgName, "Context cancelled before load test could be carried out.")
					return
				default:
					fn(ctx, client)
				}
			}
			continue
		}

		select {
		case c.loadTestScheduler.fnChan <- func(ctx context.Context) {
			select {
			case <-ctx.Done(): // This case is executed if ctx is cancelled
				c.logger.Error(pkgName, "Context cancelled before load test could be carried out.")
				return
			default:
				fn(ctx, client)
			}
		}:

		default:
			c.stats.Count1(pkgName, metricError, client.getTags(tagFunctionQueueLoadTest))
			c.logger.Error(pkgName, "load test queue is full (current queue size: %d), dropping load test request", len(c.loadTestScheduler.fnChan))
		}
	}
}

func logHystrixError(connector *connectorImpl, err error) {
	if err == nil || !strings.Contains(err.Error(), "hystrix") {
		return
	}

	connector.stats.Count1(pkgName, metricError, connector.client.getTags(tagHystrixError))
	switch e := errors.Cause(err); e {
	case hystrix.ErrTimeout:
		// handle timeout error
		connector.stats.Count1(pkgName, metricError, connector.client.getTags(tagHystrixTimeout))
		connector.logger.Warn(pkgName, "hystrix timeout error: %s", err)
	case hystrix.ErrCircuitOpen:
		// handle circuit open error
		connector.stats.Count1(pkgName, metricError, connector.client.getTags(tagHystrixCircuitOpen))
		connector.logger.Warn(pkgName, "hystrix circuit open error: %s", err)
	case hystrix.ErrMaxConcurrency:
		// handle max concurrency error
		connector.stats.Count1(pkgName, metricError, connector.client.getTags(tagHystrixMaxConcurrency))
		connector.logger.Warn(pkgName, "hystrix max concurrency error: %s", err)
	default:
		// handle other hystrix errors
		connector.logger.Warn(pkgName, "hystrix error: %s", err)
	}
}

// Do sends a redis command to a read and write enabled node
func (c *connectorImpl) Do(ctx context.Context, cmdName string, args ...interface{}) (interface{}, error) {
	if c.client.config.IgnoreReadOnly {
		readonly, _ := c.client.ifCommandReadonly(cmdName)
		if readonly {
			return c.client.Do(ctx, cmdName, args...)
		}
	}
	c.queueLoadTest(func(ctx context.Context, client *clientImpl) {
		_, _ = client.Do(ctx, cmdName, args...)
	})

	value, err := c.client.Do(ctx, cmdName, args...)
	logHystrixError(c, err)
	return value, err
}

// DoReadOnly doesn't only execute cmds on a read only node, it's the same function as Do
// Keeping this function for backward compatibility
func (c *connectorImpl) DoReadOnly(ctx context.Context, cmdName string, args ...interface{}) (interface{}, error) {
	if c.client.config.IgnoreReadOnly {
		readonly, _ := c.client.ifCommandReadonly(cmdName)
		if readonly {
			return c.client.DoReadOnly(ctx, cmdName, args...)
		}
	}
	c.queueLoadTest(func(ctx context.Context, client *clientImpl) {
		_, _ = client.DoReadOnly(ctx, cmdName, args...)
	})

	value, err := c.client.DoReadOnly(ctx, cmdName, args...)
	logHystrixError(c, err)
	return value, err
}

// Pipeline sends pipelined redis commands to a read and write enabled node and receives the reply and err
func (c *connectorImpl) Pipeline(ctx context.Context, argsList [][]interface{}) ([]redisapi.ReplyPair, error) {
	c.queueLoadTest(func(ctx context.Context, client *clientImpl) {
		_, _ = client.Pipeline(ctx, argsList)
	})

	value, err := c.client.Pipeline(ctx, argsList)
	logHystrixError(c, err)
	return value, err
}

// PipelineReadOnly doesn't only execute script on a read only node, it's the same function as Pipeline
// Keeping this function for backward compatibility
func (c *connectorImpl) PipelineReadOnly(ctx context.Context, argsList [][]interface{}) ([]redisapi.ReplyPair, error) {
	c.queueLoadTest(func(ctx context.Context, client *clientImpl) {
		_, _ = client.PipelineReadOnly(ctx, argsList)
	})

	value, err := c.client.PipelineReadOnly(ctx, argsList)
	logHystrixError(c, err)
	return value, err
}

// Run executes a script on a read and write enable node and receives the reply and err
func (c *connectorImpl) Run(ctx context.Context, script *redisapi.Script, keysAndArgs ...interface{}) (interface{}, error) {
	c.queueLoadTest(func(ctx context.Context, client *clientImpl) {
		_, _ = client.Run(ctx, script, keysAndArgs...)
	})
	value, err := c.client.Run(ctx, script, keysAndArgs...)
	logHystrixError(c, err)
	return value, err
}

// RunReadOnly doesn't only execute script on a read only node, it's the same function as Run
// Keeping this function for backward compatibility
func (c *connectorImpl) RunReadOnly(ctx context.Context, script *redisapi.Script, keysAndArgs ...interface{}) (interface{}, error) {
	c.queueLoadTest(func(ctx context.Context, client *clientImpl) {
		_, _ = client.RunReadOnly(ctx, script, keysAndArgs...)
	})
	value, err := c.client.RunReadOnly(ctx, script, keysAndArgs...)
	logHystrixError(c, err)
	return value, err
}

// Publish publishes to a Redis channel and returns a string or an error
func (c *connectorImpl) Publish(ctx context.Context, channelName string, value interface{}) (interface{}, error) {
	c.queueLoadTest(func(ctx context.Context, client *clientImpl) {
		_, _ = client.Publish(ctx, channelName, value)
	})
	value, err := c.client.Publish(ctx, channelName, value)
	logHystrixError(c, err)
	return value, err
}

// Subscribe subscribes to Redis channel(s) and return a SubscribeResponse and err
func (c *connectorImpl) Subscribe(ctx context.Context, chanBufferSize int, channels ...string) (*redisapi.SubscribeResponse, error) {
	c.queueLoadTest(func(ctx context.Context, client *clientImpl) {
		_, _ = client.Subscribe(ctx, chanBufferSize, channels...)
	})
	value, err := c.client.Subscribe(ctx, chanBufferSize, channels...)
	logHystrixError(c, err)
	return value, err
}

// ShutDown will stop the status reporting, close the pools and other clean up.
func (c *connectorImpl) ShutDown(ctx context.Context) {
	c.client.ShutDown(ctx)
	c.schedulerCancel()

	c.loadTestScheduler.wg.Wait()
	// cannot use queueLoadTest to shut down because we're closing the scheduler
	for _, client := range c.loadTestClients {
		go client.ShutDown(ctx)
	}
}

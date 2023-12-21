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
	"errors"
	"strings"
	"time"

	"github.com/grab/grab-redis/circuitbreaker"
	"github.com/grab/grab-redis/redisapi"
	goredis "github.com/grab/redis/v8"
)

// clientImpl stores the basic configuration of a redis client and implements Client interface
type clientImpl struct {
	wrappedClient clientWrapper
	closeChan     chan struct{}

	tags      []string
	config    *ClientConfig
	stats     StatsClient
	logger    Logger
	cbOptions []circuitbreaker.Option
	cmdCache  map[string]*goredis.CommandInfo
}

func NewClient(ctx context.Context, config *ClientConfig, options ...ClientOption) (redisapi.Client, error) {
	config.init()
	if err := config.validate(); err != nil {
		return nil, err
	}
	return newClient(ctx, config, options...)
}

func newClient(ctx context.Context, config *ClientConfig, options ...ClientOption) (*clientImpl, error) {
	c := &clientImpl{
		closeChan: make(chan struct{}),
		cbOptions: getDefaultCBOptions(),
		logger:    NewNoopLogger(),
		stats:     NewNoopStatsClient(),
	}

	// apply options to the client
	for _, opt := range options {
		opt(c)
	}

	var err error
	c.wrappedClient, err = config.createClient(c.cbOptions)
	if err != nil {
		return nil, err
	}

	c.config = config
	c.cmdCache, _ = c.wrappedClient.Command(ctx).Result()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	go c.monitorPool(reportInterval)

	return c, nil
}

func (c *clientImpl) reload(config *ClientConfig) error {
	return c.wrappedClient.reload(config, c.cbOptions)
}

func (c *clientImpl) ifCommandReadonly(name string) (bool, error) {
	if len(c.cmdCache) == 0 || c.cmdCache[name] == nil {
		return false, errors.New("no command cache")
	}
	return c.cmdCache[name].ReadOnly, nil
}

// Do sends a redis command to a read and write enabled node
func (c *clientImpl) Do(ctx context.Context, cmdName string, args ...interface{}) (interface{}, error) {
	defer c.stats.Duration(pkgName, metricElapsed, time.Now(), c.getTags(tagFunctionDo, tagCmdPrefix+cmdName)...)

	allArgs := redisapi.NewArgs(cmdName).Add(args...)

	return c.do(ctx, allArgs.Value()...)
}

// DoReadOnly doesn't only execute commands on a read only node, it's the same function as Do
func (c *clientImpl) DoReadOnly(ctx context.Context, cmdName string, args ...interface{}) (interface{}, error) {
	return c.Do(ctx, cmdName, args...)
}

// Pipeline sends pipelined redis commands to a read and write enabled node and receives the reply and err
func (c *clientImpl) Pipeline(ctx context.Context, argsList [][]interface{}) ([]redisapi.ReplyPair, error) {
	defer c.stats.Duration(pkgName, metricElapsed, time.Now(), c.getTags(tagFunctionPipeline)...)
	pipe := c.wrappedClient.Pipeline()

	ctx = context.Background()
	cmds := make([]*goredis.Cmd, len(argsList))
	for i, args := range argsList {
		cmd := goredis.NewCmd(ctx, args...)
		cmds[i] = cmd
		_ = pipe.Process(ctx, cmd)
	}

	_, _ = pipe.Exec(ctx)

	return c.getResultFromCommands(cmds)
}

// PipelineReadOnly doesn't only execute commands on a read only node, it's the same function as Pipeline
func (c *clientImpl) PipelineReadOnly(ctx context.Context, argsList [][]interface{}) ([]redisapi.ReplyPair, error) {
	return c.Pipeline(ctx, argsList)
}

// Run executes a script on a read and write enable node and receives the reply and err
func (c *clientImpl) Run(ctx context.Context, script *redisapi.Script, keysAndArgs ...interface{}) (interface{}, error) {
	defer c.stats.Duration(pkgName, metricElapsed, time.Now(), c.getTags(tagFunctionRun)...)
	// attempt to run the script
	args := script.GetHashAndArgs(keysAndArgs...)
	allArgs := redisapi.NewArgs(redisEvalSha).Add(args...)
	reply, err := c.do(ctx, allArgs.Value()...)
	if err != nil && strings.HasPrefix(err.Error(), redisErrNoScript) {
		args = script.GetScriptAndArgs(keysAndArgs...)
		allArgs = redisapi.NewArgs(redisEval).Add(args...)
		reply, err = c.do(ctx, allArgs.Value()...)
		if err != nil {
			c.stats.Count1(pkgName, metricError, c.getTags(tagFunctionRun))
			c.logger.Warn(pkgName, "Unable to run script with args %v. Error: %v\n", args, err)
		}
	}
	return reply, err
}

// RunReadOnly doesn't only execute commands on a read only node, it's the same function as Run
func (c *clientImpl) RunReadOnly(ctx context.Context, script *redisapi.Script, keysAndArgs ...interface{}) (interface{}, error) {
	return c.Run(ctx, script, keysAndArgs...)
}

// Publish publishes to a Redis channel and returns a string and error
func (c *clientImpl) Publish(ctx context.Context, channelName string, value interface{}) (interface{}, error) {
	return c.wrappedClient.Publish(context.Background(), channelName, value).Result()
}

// Subscribe subscribes to Redis channel(s) and return a SubscribeResponse and err
func (c *clientImpl) Subscribe(ctx context.Context, chanBufferSize int, channels ...string) (*redisapi.SubscribeResponse, error) {
	sub := c.wrappedClient.Subscribe(context.Background(), channels...)

	ch := sub.Channel()
	resultChan := make(chan interface{}, chanBufferSize)
	go func() {
		for msg := range ch {
			resultChan <- &redisapi.SubscribeMessage{
				Channel: msg.Channel,
				Data:    []byte(msg.Payload),
			}
		}

		// close channel
		close(resultChan)
	}()

	return &redisapi.SubscribeResponse{
		ResultChan: resultChan,
		Unsubscribe: func() {
			_ = sub.Unsubscribe(context.Background(), channels...)
		},
	}, nil
}

// ShutDown will stop the status reporting, close the pools and other clean up.
func (c *clientImpl) ShutDown(ctx context.Context) {
	if _, ok := ctx.Deadline(); !ok {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, defaultShutdownTimeout)
		ctx = ctxWithTimeout
		defer cancel()
	}
	go func() {
		if err := c.wrappedClient.Close(); err != nil {
			c.logger.Error(pkgName, "failed to close wrappedClient with error:%s", err)
		}
		close(c.closeChan)
	}()
	select {
	case <-c.closeChan:
		c.stats.Count1(pkgName, metricShutdown, c.getTags(tagTimeoutFalse))
		c.logger.Info(pkgName, "Gracefully shutdown")
	case <-ctx.Done():
		c.stats.Count1(pkgName, metricShutdown, c.getTags(tagTimeoutTrue))
		c.logger.Warn(pkgName, "ShutDown ctx Done with:%s", ctx.Err())
	}
}

func (c *clientImpl) do(ctx context.Context, args ...interface{}) (interface{}, error) {
	cmd := c.wrappedClient.Do(ctx, args...)
	return c.getResultFromCommand(cmd)
}

// monitorPool reports pool statistics like in conman/single_pool
func (c *clientImpl) monitorPool(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			poolStats := c.wrappedClient.PoolStats()
			total := poolStats.TotalConns
			active := poolStats.TotalConns - poolStats.IdleConns
			c.stats.Gauge("redis.connPool", metricActive, float64(active), c.getTags())
			c.stats.Gauge("redis.connPool", metricTotal, float64(total), c.getTags())
		case <-c.closeChan:
			return
		}
	}
}

func (c *clientImpl) getResultFromCommands(cmds []*goredis.Cmd) ([]redisapi.ReplyPair, error) {
	results := make([]redisapi.ReplyPair, len(cmds))
	var err error
	for idx, cmd := range cmds {
		results[idx].Value, results[idx].Err = c.getResultFromCommand(cmd)
		if err == nil && results[idx].Err != nil {
			err = results[idx].Err
		}
	}

	return results, err
}

func (c *clientImpl) getResultFromCommand(cmd *goredis.Cmd) (interface{}, error) {
	reply, err := cmd.Result()
	if err == goredis.Nil {
		err = nil
	}
	return reply, err
}

func (c *clientImpl) getTags(tags ...string) []string {
	var statsTags []string
	statsTags = append(statsTags, tagHostPrefix+c.config.name())
	statsTags = append(statsTags, tags...)
	return statsTags
}

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
	"crypto/tls"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/grab/grab-redis/circuitbreaker"
	goredis "gitlab.myteksi.net/dbops/Redis/v8"
)

type ConnectorConfig struct {
	Main      *ClientConfig   `json:"main"`
	LoadTests []*ClientConfig `json:"loadTests"`

	HotReload bool `json:"hotReload"`
	// ProcessAllLoadTestPackets this option is for not allowing losing packets when the channel is full when you dual write in new clutser.
	// It will increase the latency of requests.
	ProcessAllLoadTestPackets bool `json:"processAllLoadTestPackets"`
	// SchedulerWorkerNumber defines the max number of workers for load test scheduler, smaller number of workers will use less resources on the load test client.
	SchedulerWorkerNumber int `json:"schedulerWorkerNumber"`
	// SchedulerChannelSize specifies the max channel size for load test scheduler, it is used as a buffer for requests that are routing to load test client.
	// If the channel is full, the request will be abandoned if the ConnectorProcessAllLoadTestPackets is false.
	SchedulerChannelSize int `json:"schedulerChannelSize"`
	// SchedulerWorkerIdleTimeout specifies the max idle time for a worker, if the worker is idle for this time, it will be terminated.
	SchedulerWorkerIdleTimeoutInMs int `json:"schedulerWorkerIdleTimeout"`
}

func (c *ConnectorConfig) initAndValidate() error {
	c.Main.init()
	if err := c.Main.validate(); err != nil {
		return err
	}

	for _, config := range c.LoadTests {
		config.init()
	}
	for _, config := range c.LoadTests {
		if err := config.validate(); err != nil {
			return err
		}
	}

	if c.SchedulerWorkerNumber == 0 {
		c.SchedulerWorkerNumber = defaultMaxWorker
	}

	if c.SchedulerChannelSize == 0 {
		c.SchedulerChannelSize = defaultMaxChanSize
	}

	if c.SchedulerWorkerIdleTimeoutInMs == 0 {
		c.SchedulerWorkerIdleTimeoutInMs = defaultWorkerIdleTimeout
	}

	return nil
}

// ClientConfig keeps the settings to set up redis connector, for more details of those parameter, please refer to:https://wiki.grab.com/display/DBOps/Redis+Connector+Manual#RedisConnectorManual-ConfigurationParameterTable
type ClientConfig struct {
	// Redis connector mode, could be ModeCluster, ModeMasterSlaveGroup or ModeSingleHost
	ClientMode ClientMode `json:"clientMode"`

	// Addrs in format of host:port to connect to redis.
	// For ModeCluster, use a seed list of addresses of cluster nodes.
	// For ModeMasterSlaveGroup, use the master address followed by addresses of all slave nodes.
	// For ModeSingleHost, use only the single host address.
	Addrs []string `json:"addrs"`

	Username string `json:"username"`
	Password string `json:"password"`

	// Database to be selected after connecting to the server.
	// For ModeSingleHost only.
	DB int `json:"db"`

	MaxRetries          int `json:"maxRetries"`
	MinRetryBackoffInMs int `json:"minRetryBackoffInMs"`
	MaxRetryBackoffInMs int `json:"maxRetryBackoffInMs"`

	DialTimeoutInMs  int `json:"dialTimeoutInMs"`
	ReadTimeoutInMs  int `json:"readTimeoutInMs"`
	WriteTimeoutInMs int `json:"writeTimeoutInMs"`

	PoolSize               int `json:"poolSize"`
	MinIdleConns           int `json:"minIdleConns"`
	MaxIdleConns           int `json:"maxIdleConns"`
	MaxConnAgeInMs         int `json:"maxConnAgeInMs"`
	PoolTimeoutInMs        int `json:"poolTimeoutInMs"`
	IdleTimeoutInMs        int `json:"idleTimeoutInMs"`
	IdleCheckFrequencyInMs int `json:"idleCheckFrequencyInMs"`

	// TLSEnabled will set the InsecureSkipVerify flag in TLS to negotiate during Dail
	TLSEnabled bool `json:"tlsEnabled"`

	// Hystrix setting that is common to all nodes.
	// Each node has its own circuit breaker
	HystrixEnabled bool    `json:"hystrixEnabled"`
	Hystrix        Hystrix `json:"hystrix"`

	// The maximum number of retries among nodes before giving up.
	// Command is retried on network errors and MOVED/ASK redirects.
	// For ModeCluster and ModeMasterSlaveGroup only.
	MaxRedirects int `json:"maxRedirects"`

	// Read-only commands routing option.
	// For ModeCluster and ModeMasterSlaveGroup only.
	ReadMode ReadMode `json:"readMode"`

	// For dual write scenarios, this option is for only routing the non-readonly cmds to the new cluster to reduce traffic.
	// Only support ignore read-only cmds routing to the new cluster in Do method.
	// Enable this option will affect the prod Redis's request routing.
	IgnoreReadOnly bool `json:"ignoreReadOnly"`
}

func (c *ClientConfig) mode() string {
	return string(c.ClientMode)
}

func (c *ClientConfig) name() string {
	if len(c.Addrs) == 0 {
		return defaultHostAndPort
	}

	if c.ClientMode == ModeSingleHost {
		return c.Addrs[0]
	}

	sort.Strings(c.Addrs)
	return strings.Join(c.Addrs, ",")
}

func (c *ClientConfig) init() {
	if len(c.Addrs) == 0 {
		c.Addrs = []string{defaultHostAndPort}
	}

	if c.ReadMode == "" || c.ReadMode == ucmEmptyString {
		c.ReadMode = defaultReadMode
	}

	if c.Username == ucmEmptyString {
		c.Username = ""
	}

	if c.Password == ucmEmptyString {
		c.Password = ""
	}

	if c.DialTimeoutInMs == 0 {
		c.DialTimeoutInMs = defaultDialTimeoutInMs
	}

	if c.Hystrix.TimeoutInMs == 0 {
		c.Hystrix.TimeoutInMs = defaultCBTimeoutInMS
	}

	if c.Hystrix.MaxConcurrentRequests == 0 {
		c.Hystrix.MaxConcurrentRequests = defaultCBMaxConcurrent
	}

	if c.Hystrix.ErrorPercentThreshold == 0 {
		c.Hystrix.ErrorPercentThreshold = defaultCBErrPercent
	}

}

func (c *ClientConfig) validate() error {
	if len(c.Addrs) == 0 {
		return fmt.Errorf("no addrs found in config")
	}

	if !c.ClientMode.IsValid() {
		return fmt.Errorf("client mode %s is not valid", c.ClientMode)
	}

	if !c.ReadMode.IsValid() {
		return fmt.Errorf("read mode %s is not valid", c.ClientMode)
	}

	return nil
}

func (c *ClientConfig) validateReload(config *ClientConfig) error {
	if err := config.validate(); err != nil {
		return err
	}

	if c.ClientMode != config.ClientMode {
		return fmt.Errorf("client mode change is not allowed in reloading")
	}

	if c.DB != config.DB {
		return fmt.Errorf("DB change is not allowed in reloading")
	}

	if !isAddrsEquals(c.Addrs, config.Addrs) {
		return fmt.Errorf("addrs change is not allowed in reloading")
	}

	return nil
}

func (c *ClientConfig) createClient(cbOptions []circuitbreaker.Option) (clientWrapper, error) {
	switch c.ClientMode {
	default:
		return nil, fmt.Errorf("invalid client mode to init Redis client")
	case ModeCluster:
		return &clusterWrapperImpl{
			ClusterClient: goredis.NewDynamicClusterClient(c.clusterOptions(cbOptions)),
			config:        c,
		}, nil
	case ModeMasterSlaveGroup:
		return &clusterWrapperImpl{
			ClusterClient: goredis.NewDynamicClusterClient(c.masterSlaveGroupOptions(cbOptions)),
			config:        c,
		}, nil
	case ModeSingleHost:
		return &clientWrapperImpl{
			Client: goredis.NewDynamicClient(c.singleHostOptions(cbOptions)),
			config: c,
		}, nil
	}
}

func (c *ClientConfig) clusterOptions(cbOptions []circuitbreaker.Option) *goredis.ClusterOptions {
	opt := &goredis.ClusterOptions{
		Addrs:              c.Addrs,
		Username:           c.Username,
		Password:           c.Password,
		MaxRetries:         c.MaxRetries,
		MinRetryBackoff:    parseDurationInMs(c.MinRetryBackoffInMs),
		MaxRetryBackoff:    parseDurationInMs(c.MaxRetryBackoffInMs),
		DialTimeout:        parseDurationInMs(c.DialTimeoutInMs),
		ReadTimeout:        parseDurationInMs(c.ReadTimeoutInMs),
		WriteTimeout:       parseDurationInMs(c.WriteTimeoutInMs),
		PoolSize:           c.PoolSize,
		MinIdleConns:       c.MinIdleConns,
		MaxIdleConns:       c.MaxIdleConns,
		MaxConnAge:         parseDurationInMs(c.MaxConnAgeInMs),
		PoolTimeout:        parseDurationInMs(c.PoolTimeoutInMs),
		IdleTimeout:        parseDurationInMs(c.IdleTimeoutInMs),
		IdleCheckFrequency: parseDurationInMs(c.IdleCheckFrequencyInMs),
		MaxRedirects:       c.MaxRedirects,
	}

	switch c.ReadMode {
	case ModeReadFromMaster:
		opt.ReadOnly = false
	case ModeReadFromSlaves:
		opt.ReadOnly = true
	case ModeReadRandomly:
		opt.RouteRandomly = true
	case ModeReadByLatency:
		opt.RouteByLatency = true
	}

	if c.TLSEnabled {
		opt.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	if c.HystrixEnabled {
		opt.NewClient = func(opt *goredis.Options) *goredis.Client {
			cbKey := generateCBKey(opt.Addr)
			configureHystrix(cbKey, c.Hystrix)
			opt.Limiter = newLimiter(cbKey, cbOptions)
			return goredis.NewDynamicClient(opt)
		}
	}

	return opt
}

func (c *ClientConfig) masterSlaveGroupOptions(cbOptions []circuitbreaker.Option) *goredis.ClusterOptions {
	opt := c.clusterOptions(cbOptions)

	var nodes []goredis.ClusterNode
	for _, addr := range opt.Addrs {
		nodes = append(nodes, goredis.ClusterNode{ID: uuid.NewString(), Addr: addr})
	}

	opt.ClusterSlots = func(ctx context.Context) ([]goredis.ClusterSlot, error) {
		return []goredis.ClusterSlot{
			{Start: 0, End: 16383, Nodes: nodes},
		}, nil
	}

	return opt
}

func (c *ClientConfig) singleHostOptions(cbOptions []circuitbreaker.Option) *goredis.Options {
	addr := defaultHostAndPort
	if len(c.Addrs) > 0 {
		addr = c.Addrs[0]
	}

	opt := &goredis.Options{
		Addr:               addr,
		Username:           c.Username,
		Password:           c.Password,
		DB:                 c.DB,
		MaxRetries:         c.MaxRetries,
		MinRetryBackoff:    parseDurationInMs(c.MinRetryBackoffInMs),
		MaxRetryBackoff:    parseDurationInMs(c.MaxRetryBackoffInMs),
		DialTimeout:        parseDurationInMs(c.DialTimeoutInMs),
		ReadTimeout:        parseDurationInMs(c.ReadTimeoutInMs),
		WriteTimeout:       parseDurationInMs(c.WriteTimeoutInMs),
		PoolSize:           c.PoolSize,
		MinIdleConns:       c.MinIdleConns,
		MaxIdleConns:       c.MaxIdleConns,
		MaxConnAge:         parseDurationInMs(c.MaxConnAgeInMs),
		PoolTimeout:        parseDurationInMs(c.PoolTimeoutInMs),
		IdleTimeout:        parseDurationInMs(c.IdleTimeoutInMs),
		IdleCheckFrequency: parseDurationInMs(c.IdleCheckFrequencyInMs),
	}

	if c.TLSEnabled {
		opt.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	if c.HystrixEnabled {
		cbKey := generateCBKey(addr)
		configureHystrix(cbKey, c.Hystrix)
		opt.Limiter = newLimiter(cbKey, cbOptions)
	}

	return opt
}

func isAddrsEquals(addrs1 []string, addrs2 []string) bool {
	if len(addrs1) != len(addrs2) {
		return false
	}

	sort.Strings(addrs1)
	sort.Strings(addrs2)

	for i := range addrs1 {
		if addrs1[i] != addrs2[i] {
			return false
		}
	}

	return true
}

func parseDurationInMs(durationInMs int) time.Duration {
	return time.Duration(durationInMs) * time.Millisecond
}

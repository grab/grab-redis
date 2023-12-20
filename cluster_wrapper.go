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

	"github.com/grab/grab-redis/circuitbreaker"
	goredis "gitlab.myteksi.net/dbops/Redis/v8"
)

type clusterWrapperImpl struct {
	*goredis.ClusterClient
	config *ClientConfig
}

func (c *clusterWrapperImpl) reload(config *ClientConfig, cbOptions []circuitbreaker.Option) error {
	config.init()
	if err := c.config.validateReload(config); err != nil {
		return err
	}

	if c.config.Username != config.Username {
		c.config.Username = config.Username
		c.SetUsername(config.Username)
	}

	if c.config.Password != config.Password {
		c.config.Password = config.Password
		c.SetPassword(config.Password)
	}

	if c.config.MaxRetries != config.MaxRetries {
		c.config.MaxRetries = config.MaxRetries
		c.SetMaxRetries(config.MaxRetries)
	}

	if c.config.MinRetryBackoffInMs != config.MinRetryBackoffInMs {
		c.config.MinRetryBackoffInMs = config.MinRetryBackoffInMs
		c.SetMinRetryBackoff(parseDurationInMs(config.MinRetryBackoffInMs))
	}

	if c.config.MaxRetryBackoffInMs != config.MaxRetryBackoffInMs {
		c.config.MaxRetryBackoffInMs = config.MaxRetryBackoffInMs
		c.SetMaxRetryBackoff(parseDurationInMs(config.MaxRetryBackoffInMs))
	}

	if c.config.DialTimeoutInMs != config.DialTimeoutInMs {
		c.config.DialTimeoutInMs = config.DialTimeoutInMs
		c.SetDialTimeout(parseDurationInMs(config.DialTimeoutInMs))
	}

	if c.config.ReadTimeoutInMs != config.ReadTimeoutInMs {
		c.config.ReadTimeoutInMs = config.ReadTimeoutInMs
		c.SetReadTimeout(parseDurationInMs(config.ReadTimeoutInMs))
	}

	if c.config.WriteTimeoutInMs != config.WriteTimeoutInMs {
		c.config.WriteTimeoutInMs = config.WriteTimeoutInMs
		c.SetWriteTimeout(parseDurationInMs(config.WriteTimeoutInMs))
	}

	if c.config.PoolSize != config.PoolSize {
		c.config.PoolSize = config.PoolSize
		c.SetPoolSize(config.PoolSize)
	}

	if c.config.MinIdleConns != config.MinIdleConns {
		c.config.MinIdleConns = config.MinIdleConns
		c.SetMinIdleConns(config.MinIdleConns)
	}

	if c.config.MaxIdleConns != config.MaxIdleConns {
		c.config.MaxIdleConns = config.MaxIdleConns
		c.SetMaxIdleConns(config.MaxIdleConns)
	}

	if c.config.MaxConnAgeInMs != config.MaxConnAgeInMs {
		c.config.MaxConnAgeInMs = config.MaxConnAgeInMs
		c.SetMaxConnAge(parseDurationInMs(config.MaxConnAgeInMs))
	}

	if c.config.PoolTimeoutInMs != config.PoolTimeoutInMs {
		c.config.PoolTimeoutInMs = config.PoolTimeoutInMs
		c.SetPoolTimeout(parseDurationInMs(config.PoolTimeoutInMs))
	}

	if c.config.IdleTimeoutInMs != config.IdleTimeoutInMs {
		c.config.IdleTimeoutInMs = config.IdleTimeoutInMs
		c.SetIdleTimeout(parseDurationInMs(config.IdleTimeoutInMs))
	}

	if c.config.IdleCheckFrequencyInMs != config.IdleCheckFrequencyInMs {
		c.config.IdleCheckFrequencyInMs = config.IdleCheckFrequencyInMs
		c.SetIdleCheckFrequency(parseDurationInMs(config.IdleCheckFrequencyInMs))
	}

	if config.HystrixEnabled {
		// if the hystrix is enabled, we need to update the hystrix config when the hystrix settings changed, or it used to be disabled, but it is enabled now.
		if !c.config.Hystrix.Equals(config.Hystrix) || !c.config.HystrixEnabled {
			c.config.Hystrix = config.Hystrix
			_ = c.ForEachShard(c.Context(), func(ctx context.Context, client *goredis.Client) error {
				cbKey := generateCBKey(client.Options().Addr)
				reconfigureHystrix(cbKey, config.Hystrix)
				client.SetLimiter(newLimiter(cbKey, cbOptions))
				return nil
			})
		}
	} else {
		// if the hystrix is disabled, we need to remove the hystrix config
		_ = c.ForEachShard(c.Context(), func(ctx context.Context, client *goredis.Client) error {
			client.SetLimiter(nil)
			return nil
		})
	}
	c.config.HystrixEnabled = config.HystrixEnabled

	if c.config.MaxRedirects != config.MaxRedirects {
		c.config.MaxRedirects = config.MaxRedirects
		c.SetMaxRedirects(config.MaxRedirects)
	}

	c.config.IgnoreReadOnly = config.IgnoreReadOnly

	if c.config.ReadMode != config.ReadMode {
		c.config.ReadMode = config.ReadMode
		var readOnly, routeByLatency, routeRandomly bool
		switch config.ReadMode {
		case ModeReadFromMaster:
			readOnly, routeByLatency, routeRandomly = false, false, false
		case ModeReadFromSlaves:
			readOnly, routeByLatency, routeRandomly = true, false, false
		case ModeReadByLatency:
			readOnly, routeByLatency, routeRandomly = false, true, false
		case ModeReadRandomly:
			readOnly, routeByLatency, routeRandomly = false, false, true
		}
		c.SetReadOnly(readOnly)
		c.SetRouteByLatency(routeByLatency)
		c.SetRouteRandomly(routeRandomly)
	}

	return nil
}

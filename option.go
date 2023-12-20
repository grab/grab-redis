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
	"github.com/grab/grab-redis/circuitbreaker"
)

type ConnectorOption func(connector *connectorImpl)

// ConnectorStatsD specifies the ClientStatsD client
func ConnectorStatsD(ddClient StatsClient) ConnectorOption {
	return func(c *connectorImpl) {
		c.stats = ddClient
	}
}

// ConnectorLogger specifies the logger
func ConnectorLogger(logger Logger) ConnectorOption {
	return func(c *connectorImpl) {
		c.logger = logger
	}
}

// ConnectorCBOptions specifies the CB options
func ConnectorCBOptions(cbOptions []circuitbreaker.Option) ConnectorOption {
	return func(c *connectorImpl) {
		c.cbOptions = cbOptions
	}
}

// ClientOption is a functional parameter used to configure the clientImpl
type ClientOption func(client *clientImpl)

// ClientStatsD specifies the ClientStatsD client
func ClientStatsD(ddClient StatsClient) ClientOption {
	return func(c *clientImpl) {
		c.stats = ddClient
	}
}

// ClientLogger specifies the logger
func ClientLogger(logger Logger) ClientOption {
	return func(c *clientImpl) {
		c.logger = logger
	}
}

// ClientCBOptions specifies the CB options
func ClientCBOptions(cbOptions []circuitbreaker.Option) ClientOption {
	return func(c *clientImpl) {
		c.cbOptions = cbOptions
	}
}

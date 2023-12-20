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
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	singleHostAddr  string
	mainClusterAddr string
	loadTestAddr    string
)

func init() {
	singleHostAddr = os.Getenv("REDIS_SINGLE_HOST_ADDR")
	mainClusterAddr = os.Getenv("REDIS_MAIN_CLUSTER_ADDR")
	loadTestAddr = os.Getenv("REDIS_LOAD_TEST_CLUSTER_ADDR")
}

func clusterConfig() *ConnectorConfig {
	return &ConnectorConfig{
		Main: &ClientConfig{
			ClientMode:     ModeCluster,
			Addrs:          []string{mainClusterAddr},
			PoolSize:       10,
			IgnoreReadOnly: true,
		},
		LoadTests: []*ClientConfig{{
			ClientMode: ModeCluster,
			Addrs:      []string{loadTestAddr},
			PoolSize:   10,
		}},
		HotReload: true,
	}
}

func loadTestValidation() *ConnectorConfig {
	return &ConnectorConfig{
		Main: &ClientConfig{
			ClientMode: ModeCluster,
			Addrs:      []string{loadTestAddr},
			PoolSize:   10,
		},
		HotReload: true,
	}
}

// We don't test master-slave configuration here, because it's the same code as clusterConfig.
func singleHostConfig() *ConnectorConfig {
	return &ConnectorConfig{
		Main: &ClientConfig{
			ClientMode: ModeSingleHost,
			Addrs:      []string{singleHostAddr},
			PoolSize:   10,
		},
		HotReload: true,
	}
}

func TestGinkgo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "connector command")
}

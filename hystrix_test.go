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
	"encoding/json"
	"fmt"
	"time"

	"github.com/myteksi/hystrix-go/hystrix"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/grab/grab-redis/circuitbreaker"
	goredis "github.com/grab/redis/v8"
)

type fakeConfigurer struct {
	config   *ConnectorConfig
	callback func() error
}

func (c *fakeConfigurer) OnChange(callback func() error) {
	c.callback = callback
}

func (c *fakeConfigurer) Unmarshal(obj interface{}) error {
	data, err := json.Marshal(c.config)
	if err != nil {
		return fmt.Errorf("json marshal failed, err: %s", err)
	}

	err = json.Unmarshal(data, obj)
	if err != nil {
		return fmt.Errorf("json unmarshal failed, err: %s", err)
	}
	return nil
}

const (
	sleepWindowInMs       = 1
	circuitOpenBufferTime = 10 * time.Millisecond
)

func CancelledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func oneCircuitOpened(cbKeys []string) bool {
	var openCount int
	for _, key := range cbKeys {
		if circuitbreaker.IsCircuitOpen(key) {
			openCount++
		}
	}
	return openCount == 1
}

func anyCircuitOpened(cbKeys []string) bool {
	for _, key := range cbKeys {
		if circuitbreaker.IsCircuitOpen(key) {
			return true
		}
	}
	return false
}

var _ = Describe("hystrix in CLUSTER MODE", func() {
	var configurer *fakeConfigurer

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()

	BeforeEach(func() {
		config := clusterConfig()
		config.Main.Hystrix = Hystrix{
			RequestVolumeThreshold: 1,
			ErrorPercentThreshold:  1,
			SleepWindowInMs:        sleepWindowInMs,
		}
		configurer = &fakeConfigurer{config: config}
		hystrix.Flush()
	})

	AfterEach(func() {
		hystrix.Flush()
	})

	It("different condition of triggering circuit", func() {
		// init with hystrix
		configurer.config.Main.HystrixEnabled = true
		client, _ := NewDynamicConnector(context.Background(), configurer)
		defer client.ShutDown(context.Background())
		// for cluster, the cbKey might not equal to config.Main.Addrs[0], therefore we need to find the cbKey via ForEachShard
		var cbKeys []string
		_ = client.(*connectorImpl).client.wrappedClient.(*clusterWrapperImpl).ForEachShard(context.Background(), func(ctx context.Context, client *goredis.Client) error {
			cbKeys = append(cbKeys, generateCBKey(client.Options().Addr))
			return nil
		})
		Expect(anyCircuitOpened(cbKeys)).To(Equal(false))

		// Redis error should not be counted as CB error
		_, err := client.DoReadOnly(context.Background(), "invalid_cmd")
		time.Sleep(circuitOpenBufferTime)
		Expect(err).To(HaveOccurred())
		Expect(anyCircuitOpened(cbKeys)).To(Equal(false))

		// context.DeadlineExceeded should be counted as CB error
		_, err = client.DoReadOnly(timeoutCtx, "get", "key")
		time.Sleep(circuitOpenBufferTime)
		Expect(err).To(Equal(context.DeadlineExceeded))
		Expect(oneCircuitOpened(cbKeys)).To(Equal(true))

		// wait for circuit to allow request
		time.Sleep(sleepWindowInMs * time.Millisecond)

		// context.Canceled should not be counted as CB error

		_, err = client.DoReadOnly(CancelledContext(), "get", "key")
		time.Sleep(circuitOpenBufferTime)
		Expect(err).To(Equal(context.Canceled))
		Expect(anyCircuitOpened(cbKeys)).To(Equal(false))
	})

	It("dynamically enable hystrix works", func() {
		// init without hystrix
		client, _ := NewDynamicConnector(context.Background(), configurer)
		defer client.ShutDown(context.Background())
		// for cluster, the cbKey might not equal to config.Main.Addrs[0], therefore we need to find the cbKey via ForEachShard
		var cbKeys []string
		_ = client.(*connectorImpl).client.wrappedClient.(*clusterWrapperImpl).ForEachShard(context.Background(), func(ctx context.Context, client *goredis.Client) error {
			cbKeys = append(cbKeys, generateCBKey(client.Options().Addr))
			return nil
		})
		Expect(anyCircuitOpened(cbKeys)).To(Equal(false))

		_, err := client.DoReadOnly(timeoutCtx, "get", "key")
		time.Sleep(circuitOpenBufferTime)
		Expect(err).To(Equal(context.DeadlineExceeded))
		Expect(anyCircuitOpened(cbKeys)).To(Equal(false))

		// enable hystrix
		configurer.config.Main.HystrixEnabled = true
		err = configurer.callback()
		Expect(err).NotTo(HaveOccurred())

		_, err = client.DoReadOnly(timeoutCtx, "get", "key")
		time.Sleep(circuitOpenBufferTime)
		Expect(err).To(Equal(context.DeadlineExceeded))
		Expect(oneCircuitOpened(cbKeys)).To(Equal(true))
	})
})

var _ = Describe("hystrix in NON CLUSTER MODE", func() {
	var configurer *fakeConfigurer
	var cbKey string

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()

	BeforeEach(func() {
		config := singleHostConfig()
		config.Main.Hystrix = Hystrix{
			RequestVolumeThreshold: 1,
			ErrorPercentThreshold:  1,
			SleepWindowInMs:        sleepWindowInMs,
		}
		configurer = &fakeConfigurer{config: config}
		cbKey = generateCBKey(config.Main.Addrs[0])
		hystrix.Flush()
	})
	AfterEach(func() {
		hystrix.Flush()
	})

	It("different condition of triggering circuit", func() {
		// init with hystrix
		configurer.config.Main.HystrixEnabled = true
		client, _ := NewDynamicConnector(context.Background(), configurer)
		defer client.ShutDown(context.Background())
		Expect(circuitbreaker.IsCircuitOpen(cbKey)).To(Equal(false))

		// Redis error should not be counted as CB error
		_, err := client.DoReadOnly(context.Background(), "invalid_cmd")
		time.Sleep(circuitOpenBufferTime)
		Expect(err).To(HaveOccurred())
		Expect(circuitbreaker.IsCircuitOpen(cbKey)).To(Equal(false))

		// context.DeadlineExceeded should be counted as CB error
		_, err = client.DoReadOnly(timeoutCtx, "get", "key")
		time.Sleep(circuitOpenBufferTime)
		Expect(err).To(Equal(context.DeadlineExceeded))
		Expect(circuitbreaker.IsCircuitOpen(cbKey)).To(Equal(true))

		// wait for circuit to allow request
		time.Sleep(sleepWindowInMs * time.Millisecond)

		// context.Canceled should not be counted as CB error
		_, err = client.DoReadOnly(CancelledContext(), "get", "key")
		time.Sleep(circuitOpenBufferTime)
		Expect(err).To(Equal(context.Canceled))
		Expect(circuitbreaker.IsCircuitOpen(cbKey)).To(Equal(false))
	})

	It("dynamically enable hystrix works", func() {
		// init without hystrix
		client, _ := NewDynamicConnector(context.Background(), configurer)
		defer client.ShutDown(context.Background())
		Expect(circuitbreaker.IsCircuitOpen(cbKey)).To(Equal(false))

		_, err := client.DoReadOnly(timeoutCtx, "get", "key")
		time.Sleep(circuitOpenBufferTime)
		Expect(err).To(Equal(context.DeadlineExceeded))
		Expect(circuitbreaker.IsCircuitOpen(cbKey)).To(Equal(false))

		// enable hystrix
		configurer.config.Main.HystrixEnabled = true
		err = configurer.callback()
		Expect(err).NotTo(HaveOccurred())

		_, err = client.DoReadOnly(timeoutCtx, "get", "key")
		time.Sleep(circuitOpenBufferTime)
		Expect(err).To(Equal(context.DeadlineExceeded))
		Expect(circuitbreaker.IsCircuitOpen(cbKey)).To(Equal(true))
	})
})

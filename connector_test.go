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
	"github.com/grab/grab-redis/redisapi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Connector CLUSTER ON test", func() {
	var client redisapi.Client

	BeforeEach(func() {
		client, _ = NewStaticConnector(context.Background(), clusterConfig())
		_, err := client.Do(context.Background(), "flushall")
		Expect(err).NotTo(HaveOccurred())
	})

	It("supports context", func() {
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		cancel()
		_, err := client.Do(ctx, "PING")
		Expect(err).To(MatchError("context canceled"))
		client.ShutDown(context.Background())
	})

	It("should ping", func() {
		reply, err := client.Do(context.Background(), "PING")
		Expect(err).NotTo(HaveOccurred())
		Expect(reply).To(Equal("PONG"))
		client.ShutDown(context.Background())
	})

	It("should set", func() {
		reply, err := client.Do(context.Background(), "SET", "key", "value")
		Expect(err).NotTo(HaveOccurred())
		Expect(reply).To(Equal("OK"))
		client.ShutDown(context.Background())
	})

	It("should close", func() {
		client.ShutDown(context.Background())
		_, err := client.Do(context.Background(), "PING")
		Expect(err).To(MatchError("redis: client is closed"))
	})

})

var _ = Describe("Connector CLUSTER off test", func() {
	var client redisapi.Client

	BeforeEach(func() {
		client, _ = NewStaticConnector(context.Background(), singleHostConfig())
		_, err := client.Do(context.Background(), "flushall")
		Expect(err).NotTo(HaveOccurred())
	})

	It("supports context", func() {
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		cancel()
		_, err := client.Do(ctx, "PING")
		Expect(err).To(MatchError("context canceled"))
		client.ShutDown(context.Background())
	})

	It("should ping", func() {
		reply, err := client.Do(context.Background(), "PING")
		Expect(err).NotTo(HaveOccurred())
		Expect(reply).To(Equal("PONG"))
		client.ShutDown(context.Background())
	})

	It("should set", func() {
		reply, err := client.Do(context.Background(), "SET", "key", "value")
		Expect(err).NotTo(HaveOccurred())
		Expect(reply).To(Equal("OK"))
		client.ShutDown(context.Background())
	})

	It("should close", func() {
		client.ShutDown(context.Background())
		_, err := client.Do(context.Background(), "PING")
		Expect(err).To(MatchError("redis: client is closed"))
	})
})

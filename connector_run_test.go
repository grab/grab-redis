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

var _ = Describe("Run CLUSTER ON", func() {
	var client redisapi.Client

	BeforeEach(func() {
		config := clusterConfig()
		client, _ = NewStaticConnector(context.Background(), config)
		_, err := client.Do(context.Background(), "flushall")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		client.ShutDown(context.Background())
	})

	It("test eval", func() {
		script := redisapi.NewScript(0, "return ARGV[1]")
		script2 := redisapi.NewScript(2, "return ARGV[1]")
		run, _ := client.Run(context.Background(), script, "hello")
		Expect(run).To(Equal("hello"))
		run, _ = client.RunReadOnly(context.Background(), script2, "hello", "hello", "world")
		Expect(run).To(Equal("world"))
	})
})

var _ = Describe("Run CLUSTER OFF", func() {
	var client redisapi.Client

	BeforeEach(func() {
		config := singleHostConfig()
		client, _ = NewStaticConnector(context.Background(), config)
		_, err := client.Do(context.Background(), "flushall")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		client.ShutDown(context.Background())
	})

	It("test eval", func() {
		script := redisapi.NewScript(0, "return ARGV[1]")
		script2 := redisapi.NewScript(2, "return ARGV[1]")
		run, _ := client.Run(context.Background(), script, "hello")
		Expect(run).To(Equal("hello"))
		run, _ = client.RunReadOnly(context.Background(), script2, "hello", "hello", "world")
		Expect(run).To(Equal("world"))
	})
})

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
	"strconv"
	"time"

	"github.com/grab/grab-redis/redisapi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cmd CLUSTER ON", func() {
	var client redisapi.Client

	BeforeEach(func() {
		config := clusterConfig()
		client, _ = NewStaticConnector(context.Background(), config)
	})

	AfterEach(func() {
		client.ShutDown(context.Background())
	})

	It("implements Stringer", func() {
		set, _ := client.Do(context.Background(), "set", "foo", "bar")
		Expect(set).To(Equal("OK"))

		get, _ := client.DoReadOnly(context.Background(), "get", "foo")
		Expect(get).To(Equal("bar"))
	})

	It("has val/err", func() {
		set, err := client.Do(context.Background(), "set", "key", "hello")
		Expect(err).NotTo(HaveOccurred())
		Expect(set).To(Equal("OK"))

		get, err := client.DoReadOnly(context.Background(), "get", "key")
		Expect(err).NotTo(HaveOccurred())
		Expect(get).To(Equal("hello"))
	})

	It("supports float64", func() {
		f := float32(66.97)

		_, err := client.Do(context.Background(), "set", "float_key", f)
		Expect(err).NotTo(HaveOccurred())

		val, err := client.DoReadOnly(context.Background(), "get", "float_key")
		Expect(err).NotTo(HaveOccurred())

		valStr, ok := val.(string)
		Expect(ok).To(Equal(true))

		val1, _ := strconv.ParseFloat(valStr, 32)
		val2 := float32(val1)
		Expect(val2).To(Equal(f))
	})

	It("supports time.Time", func() {
		tm := time.Date(2019, 1, 1, 9, 45, 10, 222125, time.UTC)

		_, err := client.Do(context.Background(), "set", "time_key", tm)
		Expect(err).NotTo(HaveOccurred())

		s, err := client.DoReadOnly(context.Background(), "get", "time_key")
		Expect(err).NotTo(HaveOccurred())
		Expect(s).To(Equal("2019-01-01T09:45:10.000222125Z"))
	})
})

var _ = Describe("Cmd CLUSTER OFF", func() {
	var client redisapi.Client

	BeforeEach(func() {
		config := singleHostConfig()
		client, _ = NewStaticConnector(context.Background(), config)
	})

	AfterEach(func() {
		client.ShutDown(context.Background())
	})

	It("implements Stringer", func() {
		set, _ := client.Do(context.Background(), "set", "foo", "bar")
		Expect(set).To(Equal("OK"))

		get, _ := client.Do(context.Background(), "get", "foo")
		Expect(get).To(Equal("bar"))
	})

	It("has val/err", func() {
		set, err := client.Do(context.Background(), "set", "key", "hello")
		Expect(err).NotTo(HaveOccurred())
		Expect(set).To(Equal("OK"))

		get, err := client.Do(context.Background(), "get", "key")
		Expect(err).NotTo(HaveOccurred())
		Expect(get).To(Equal("hello"))
	})

	It("supports float64", func() {
		f := float32(66.97)

		_, err := client.Do(context.Background(), "set", "float_key", f)
		Expect(err).NotTo(HaveOccurred())

		val, err := client.DoReadOnly(context.Background(), "get", "float_key")
		Expect(err).NotTo(HaveOccurred())

		valStr, ok := val.(string)
		Expect(ok).To(Equal(true))

		val1, _ := strconv.ParseFloat(valStr, 32)
		val2 := float32(val1)
		Expect(val2).To(Equal(f))
	})

	It("supports time.Time", func() {
		tm := time.Date(2019, 1, 1, 9, 45, 10, 222125, time.UTC)

		_, err := client.Do(context.Background(), "set", "time_key", tm)
		Expect(err).NotTo(HaveOccurred())

		s, err := client.Do(context.Background(), "get", "time_key")
		Expect(err).NotTo(HaveOccurred())
		Expect(s).To(Equal("2019-01-01T09:45:10.000222125Z"))
	})
})

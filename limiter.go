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
	"fmt"

	cb "github.com/grab/grab-redis/circuitbreaker"
	"github.com/myteksi/hystrix-go/hystrix"
	goredis "gitlab.myteksi.net/dbops/Redis/v8"
)

// We need to use a separate circuit breaker for each redis instance, we implemented and modified the limiter that Go-Redis provides to support per node cb.
type limiter struct {
	ctx       context.Context
	key       string
	cbOptions []cb.Option
}

func newLimiter(key string, cbOptions []cb.Option) *limiter {
	l := &limiter{
		ctx:       context.Background(),
		key:       key,
		cbOptions: cbOptions,
	}
	return l
}

func (l limiter) Allow() error {
	if !cb.AllowRequest(l.key) {
		return hystrix.ErrCircuitOpen
	}
	return nil
}

func (l limiter) Execute(fn func() error) error {
	return cb.Do(l.ctx, l.key, fn, l.cbOptions...)
}

// ReportResult do nothing as Execute will handle the error reporting.
func (l limiter) ReportResult(result error) {
}

func generateCBKey(addr string) string {
	return fmt.Sprintf("redis_%s", addr)
}

func getDefaultCBOptions() []cb.Option {
	return []cb.Option{
		cb.WithUserErrorHandler(func(err error) (nonThreat bool, errOut error) {
			return !shouldCountAsCBError(err), err
		}),
	}
}

func shouldCountAsCBError(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, context.Canceled) {
		return false
	}

	// Count the readonly error as cb error
	if goredis.IsReadOnlyError(err) {
		return true
	}

	// Redis error should not be counted as CB error
	if _, ok := err.(goredis.Error); ok {
		return false
	}

	return true
}

func configureHystrix(key string, setting Hystrix) {
	builder := cb.NewCircuitBuilder(key).
		WithTimeout(setting.TimeoutInMs).
		WithMaxConcurrentRequests(setting.MaxConcurrentRequests).
		WithRequestVolumeThreshold(setting.RequestVolumeThreshold).
		WithErrorPercentageThreshold(setting.ErrorPercentThreshold).
		WithSleepWindow(setting.SleepWindowInMs).
		WithQueueSize(setting.QueueSizeRejectionThreshold)

	cb.ConfigureCircuit(builder.Build())
}

func reconfigureHystrix(key string, setting Hystrix) {
	configureHystrix(key, setting)
	hystrix.Flush() // remove any existing circuit to apply new setting
}

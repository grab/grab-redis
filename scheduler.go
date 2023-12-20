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
	"math"
	"sync"
	"time"

	"go.uber.org/atomic"
)

type schedulerOptions struct {
	maxChanSize       int
	maxWorker         int
	workerIdleTimeout time.Duration
}

func (s *schedulerOptions) normalise() {
	if s.maxChanSize <= 0 {
		s.maxChanSize = defaultMaxChanSize
	}

	if s.maxWorker <= 0 {
		s.maxWorker = defaultMaxWorker
	}

	if s.workerIdleTimeout <= 0 {
		s.workerIdleTimeout = defaultWorkerIdleTimeout
	}
}

type scheduler struct {
	fnChan            chan func(ctx context.Context)
	maxWorker         int
	workerIdleTimeout time.Duration

	latencies *latencies
	numWorker *atomic.Int64

	wg *sync.WaitGroup
}

func (s *scheduler) spawnWorker(ctx context.Context) {
	s.numWorker.Inc()
	defer s.numWorker.Dec()
	defer s.wg.Done()
	idleTimeoutTicker := time.NewTicker(s.workerIdleTimeout)
	defer idleTimeoutTicker.Stop()

	for {
		select {
		case fn := <-s.fnChan:
			start := time.Now()
			fn(ctx)
			elapsed := time.Since(start)
			s.latencies.Add(elapsed.Nanoseconds())

			idleTimeoutTicker.Reset(s.workerIdleTimeout)
		case <-idleTimeoutTicker.C:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (s *scheduler) start(ctx context.Context) {
	backlogTicker := time.NewTicker(100 * time.Millisecond)
	defer backlogTicker.Stop()

	for {
		select {
		case <-backlogTicker.C:
			backlog := len(s.fnChan)
			if backlog == 0 {
				continue
			}

			need := int64(1)
			avgLatency := s.latencies.Average() // nanoseconds
			if avgLatency > 0 {
				need = int64(math.Ceil(float64(backlog) * (avgLatency / 1e9)))
			}

			for i := int64(0); i < need && s.numWorker.Load() < int64(s.maxWorker); i++ {
				s.wg.Add(1)
				go s.spawnWorker(ctx)
			}
		case <-ctx.Done():
			s.wg.Wait()
			close(s.fnChan)
			return
		}
	}
}

func newScheduler(options *schedulerOptions) *scheduler {
	options.normalise()
	return &scheduler{
		fnChan:            make(chan func(context.Context), options.maxChanSize),
		maxWorker:         options.maxWorker,
		workerIdleTimeout: options.workerIdleTimeout,
		latencies:         newLatencies(1000),
		numWorker:         atomic.NewInt64(int64(0)),
		wg:                &sync.WaitGroup{},
	}
}

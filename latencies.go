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

// Script taken from gandalf

package redis

import (
	"sync/atomic"
)

type latencies struct {
	values []int64
	index  int64
	added  int64
}

func newLatencies(l int) *latencies {
	return &latencies{
		values: make([]int64, l),
		index:  -1,
	}
}

func (l *latencies) Add(value int64) {
	index := atomic.AddInt64(&l.index, 1)
	index = index % int64(len(l.values))

	// the guarantee here is that the old value is replaced with a newer one,
	// so we ignore whether this might have overridden an even newer value
	atomic.StoreInt64(&l.values[index], value)

	// after the store, there are at least `added` number of elements
	atomic.AddInt64(&l.added, 1)
}

func (l *latencies) Average() float64 {
	var total int64

	// we cannot use index since index is incremented before the swap occurs
	added := int(atomic.LoadInt64(&l.added))

	var length int
	if added < len(l.values) {
		length = added
	} else {
		length = len(l.values)
	}

	if length <= 0 {
		return 0.0
	}

	// we are not concerned whether all values are atomically retrieved at once since we are sampling an estimate
	for i := 0; i < length; i++ {
		total += atomic.LoadInt64(&l.values[i])
	}

	return float64(total) / float64(length)
}

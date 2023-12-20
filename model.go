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

type ClientMode string

const (
	ModeCluster          ClientMode = "cluster"
	ModeMasterSlaveGroup ClientMode = "masterSlaveGroup"
	ModeSingleHost       ClientMode = "singleHost"
)

func (m ClientMode) In(modes ...ClientMode) bool {
	for _, mode := range modes {
		if m == mode {
			return true
		}
	}

	return false
}

func (m ClientMode) IsValid() bool {
	return m.In(ModeCluster, ModeMasterSlaveGroup, ModeSingleHost)
}

type ReadMode string

const (
	ModeReadFromMaster ReadMode = "readFromMaster"
	ModeReadFromSlaves ReadMode = "readFromSlaves"
	ModeReadRandomly   ReadMode = "readRandomly"
	ModeReadByLatency  ReadMode = "readByLatency"
)

func (m ReadMode) In(modes ...ReadMode) bool {
	for _, mode := range modes {
		if m == mode {
			return true
		}
	}

	return false
}

func (m ReadMode) IsValid() bool {
	return m.In(ModeReadFromMaster, ModeReadFromSlaves, ModeReadRandomly, ModeReadByLatency)
}

// Hystrix circuit breaker setting
type Hystrix struct {
	// TimeoutInMs is how long to wait for command to complete, in milliseconds
	TimeoutInMs int `json:"timeoutInMs"`
	// MaxConcurrentRequests is how many commands of the same type can run at the same time
	MaxConcurrentRequests int `json:"maxConcurrentRequests"`
	// RequestVolumeThreshold is the minimum number of requests needed before a circuit can be tripped due to health
	RequestVolumeThreshold int `json:"requestVolumeThreshold"`
	// ErrorPercentThreshold causes circuits to open once the rolling measure of errors exceeds this percent of requests
	ErrorPercentThreshold int `json:"errorPercentThreshold"`
	// SleepWindowInMs is how long, in milliseconds, to wait after a circuit opens before testing for recovery
	SleepWindowInMs int `json:"sleepWindowInMs"`
	// QueueSizeRejectionThreshold reject requests when the queue size exceeds the given limit
	QueueSizeRejectionThreshold int `json:"queueSizeRejectionThreshold"`
}

func (h Hystrix) Equals(hystrix Hystrix) bool {
	return h.TimeoutInMs == hystrix.TimeoutInMs &&
		h.MaxConcurrentRequests == hystrix.MaxConcurrentRequests &&
		h.RequestVolumeThreshold == hystrix.RequestVolumeThreshold &&
		h.ErrorPercentThreshold == hystrix.ErrorPercentThreshold &&
		h.SleepWindowInMs == hystrix.SleepWindowInMs &&
		h.QueueSizeRejectionThreshold == hystrix.QueueSizeRejectionThreshold
}

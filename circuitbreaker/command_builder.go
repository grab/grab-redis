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

package circuitbreaker

import (
	"time"

	"github.com/myteksi/hystrix-go/hystrix"
)

// CommandBuilder builder for constructing new command

type CommandBuilder struct {
	commandName            string
	timeout                int
	maxConcurrentRequests  int
	requestVolumeThreshold int
	sleepWindow            int
	errorPercentThreshold  int
	// for more details refer - https://github.com/Netflix/Hystrix/wiki/Configuration#maxqueuesize
	queueSizeRejectionThreshold *int
	// group a number of command (circuit name) together, useful for defining ownership/alerts/monitoring
	// ref: https://github.com/Netflix/Hystrix/wiki/How-To-Use#command-group
	commandGroup string
}

// New Create new command
func New(commandName string) *CommandBuilder {
	return &CommandBuilder{
		commandName:                 commandName,
		commandGroup:                "",
		timeout:                     hystrix.DefaultTimeout,
		maxConcurrentRequests:       hystrix.DefaultMaxConcurrent,
		requestVolumeThreshold:      hystrix.DefaultVolumeThreshold,
		sleepWindow:                 hystrix.DefaultSleepWindow,
		errorPercentThreshold:       hystrix.DefaultErrorPercentThreshold,
		queueSizeRejectionThreshold: nil, // will init later on build
	}
}

// WithCommandGroup modify commandGroup
func (cb *CommandBuilder) WithCommandGroup(commandGroup string) *CommandBuilder {
	if commandGroup != "" {
		cb.commandGroup = commandGroup
	}
	return cb
}

// WithTimeout modify timeout
func (cb *CommandBuilder) WithTimeout(timeoutInMs int) *CommandBuilder {
	if timeoutInMs > 0 {
		cb.timeout = timeoutInMs
	}
	return cb
}

// WithMaxConcurrentRequests modify max concurrent requests
// if not already set, this will also set the queue size as 5 times the max concurrent requests
func (cb *CommandBuilder) WithMaxConcurrentRequests(maxConcurrentRequests int) *CommandBuilder {
	if maxConcurrentRequests > 0 {
		cb.maxConcurrentRequests = maxConcurrentRequests
	}
	return cb
}

// WithErrorPercentageThreshold modify error percentage threshold
func (cb *CommandBuilder) WithErrorPercentageThreshold(errPercentThreshold int) *CommandBuilder {
	if errPercentThreshold > 0 {
		cb.errorPercentThreshold = errPercentThreshold
	}
	return cb
}

// WithRequestVolumeThreshold modify request volume threshold
func (cb *CommandBuilder) WithRequestVolumeThreshold(requestVolThreshold int) *CommandBuilder {
	if requestVolThreshold > 0 {
		cb.requestVolumeThreshold = requestVolThreshold
	}
	return cb
}

// WithSleepWindow modify sleep window
func (cb *CommandBuilder) WithSleepWindow(sleepWindow int) *CommandBuilder {
	if sleepWindow > 0 {
		cb.sleepWindow = sleepWindow
	}
	return cb
}

// WithQueueSize modify queue size
func (cb *CommandBuilder) WithQueueSize(queueSize int) *CommandBuilder {
	if queueSize == 0 {
		zeroQueueSize := 0
		cb.queueSizeRejectionThreshold = &zeroQueueSize
	} else if queueSize > 0 {
		cb.queueSizeRejectionThreshold = &queueSize
	}
	return cb
}

// Build the command setting, Use hystrix.Initialize for setup
func (cb *CommandBuilder) Build() *hystrix.Settings {

	// if value is not set, we'll use default 5x of max concurrent
	if cb.queueSizeRejectionThreshold == nil {
		queueSize := 5 * cb.maxConcurrentRequests
		cb.queueSizeRejectionThreshold = &queueSize
	}

	return &hystrix.Settings{
		CommandName:                 cb.commandName,
		CommandGroup:                cb.commandGroup,
		Timeout:                     time.Duration(cb.timeout) * time.Millisecond,
		MaxConcurrentRequests:       cb.maxConcurrentRequests,
		ErrorPercentThreshold:       cb.errorPercentThreshold,
		RequestVolumeThreshold:      uint64(cb.requestVolumeThreshold),
		SleepWindow:                 time.Duration(cb.sleepWindow) * time.Millisecond,
		QueueSizeRejectionThreshold: *cb.queueSizeRejectionThreshold,
	}
}

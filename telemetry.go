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

import "time"

// Logger defines the interface for logging different levels of messages.
type Logger interface {
	// Error logs an error message with the given format and arguments.
	Error(packageName string, format string, args ...interface{})

	// Warn logs a warning message with the given format and arguments.
	Warn(packageName string, format string, args ...interface{})

	// Info logs an informational message with the given format and arguments.
	Info(packageName string, format string, args ...interface{})

	// ServiceDown logs a service down warning with the service name and error.
	ServiceDown(name string, err error)

	// ContextError logs an error message when a context operation fails.
	ContextError(name string, err error)
}

// StatsClient defines the interface for interacting with a stats system like StatsD.
type StatsClient interface {
	// Count1 is a convenience method incrementing the counter by 1 for the given metric.
	Count1(pkgName string, metric string, tags ...[]string)

	// Gauge records a value for the given metric, typically a value at a point in time.
	Gauge(name string, metric string, value float64, tags []string)

	// Duration records the duration of an event for the given metric.
	Duration(name string, elapsed string, now time.Time, tags ...string)
}

// NoopLogger is an implementation of Logger that does not log any messages.
type NoopLogger struct{}

// Error does nothing.
func (l *NoopLogger) Error(packageName string, format string, args ...interface{}) {}

// Warn does nothing.
func (l *NoopLogger) Warn(packageName string, format string, args ...interface{}) {}

// Info does nothing.
func (l *NoopLogger) Info(packageName string, format string, args ...interface{}) {}

// ServiceDown does nothing.
func (l *NoopLogger) ServiceDown(name string, err error) {}

// ContextError does nothing.
func (l *NoopLogger) ContextError(name string, err error) {}

// NewNoopLogger returns a new instance of a logger that does not perform any logging.
func NewNoopLogger() Logger {
	return &NoopLogger{}
}

// NoopStatsClient is an implementation of StatsClient that does not record any statistics.
type NoopStatsClient struct{}

// Count1 does nothing.
func (n *NoopStatsClient) Count1(pkgName string, metric string, tags ...[]string) {}

// Gauge does nothing.
func (n *NoopStatsClient) Gauge(name string, metric string, value float64, tags []string) {}

// Duration does nothing.
func (n *NoopStatsClient) Duration(name string, elapsed string, now time.Time, tags ...string) {}

// NewNoopStatsClient returns a new instance of StatsClient that does not record any statistics.
func NewNoopStatsClient() StatsClient {
	return &NoopStatsClient{}
}

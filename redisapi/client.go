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

package redisapi

import (
	"context"
)

type Client interface {
	Doer
	Pipeliner
	Runner
	Publisher
	Subscriber
	Closer
}

// Doer interface defines something send single redis commands
type Doer interface {
	// Do sends a redis command to a read and write enabled node
	Do(ctx context.Context, cmdName string, args ...interface{}) (interface{}, error)

	// DoReadOnly doesn't only execute script on a read only node, it's the same function as Do
	// Keeping this function for backward compatibility
	DoReadOnly(ctx context.Context, cmdName string, args ...interface{}) (interface{}, error)
}

// Pipeliner interface defines something that can send pipelined redis requested
type Pipeliner interface {
	// Pipeline sends pipelined redis commands to a read and write enabled node and receives the reply and err
	Pipeline(ctx context.Context, args [][]interface{}) ([]ReplyPair, error)

	// PipelineReadOnly doesn't only execute script on a read only node, it's the same function as Pipeline
	// Keeping this function for backward compatibility
	PipelineReadOnly(ctx context.Context, args [][]interface{}) ([]ReplyPair, error)
}

// Runner interface defines something that can run redis LUA scripts
type Runner interface {
	// Run executes a script on a read and write enable node and receives the reply and err
	Run(ctx context.Context, script *Script, keysAndArgs ...interface{}) (interface{}, error)

	// RunReadOnly doesn't only execute script on a read only node, it's the same function as Run
	// Keeping this function for backward compatibility
	RunReadOnly(ctx context.Context, script *Script, keysAndArgs ...interface{}) (interface{}, error)
}

// Publisher interface defines something that can perform redis publish
type Publisher interface {
	// Publish publishes to a Redis channel and returns an error
	Publish(ctx context.Context, channelName string, value interface{}) (interface{}, error)
}

// UnsubscribeFunc tells a connection to cancel all it's subscriptions
type UnsubscribeFunc func()

// SubscribeMessage message notification from the pubsub channel
type SubscribeMessage struct {
	// The originating channel.
	Channel string

	// The message data.
	Data []byte
}

// SubscribeResponse encapsulates the response of a subscribed call
// ResultChan contains all the messages received from the subscription
// Unsubscribe can be used to terminate the subscription
type SubscribeResponse struct {
	// ResultChan returns either a SubscribeMessage or an error
	ResultChan  <-chan interface{}
	Unsubscribe UnsubscribeFunc
}

// Subscriber interface defines something that can perform redis subscribe
type Subscriber interface {
	// Subscribe subscribes to Redis channel(s) and return a SubscribeResponse and err
	Subscribe(ctx context.Context, bufferSize int, channels ...string) (response *SubscribeResponse, err error)
}

// ReplyPair is the general struct for response of a single redis command
type ReplyPair struct {
	Value interface{}
	Err   error
}

// Closer interface defines something that can close the redis connector
type Closer interface {
	// ShutDown stops the status reporting, close the pools and other clean up.
	ShutDown(ctx context.Context)
}

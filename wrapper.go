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

	cb "github.com/grab/grab-redis/circuitbreaker"
	goredis "github.com/grab/redis/v8"
)

// wrap goRedis calls (to allow mock injection)
//
//go:generate mockery -name clientWrapper -inpkg -case=underscore -testonly
type clientWrapper interface {
	redisWrapper
	reload(config *ClientConfig, cbOption []cb.Option) error
}

type redisWrapper interface {
	Do(ctx context.Context, args ...interface{}) *goredis.Cmd
	Process(ctx context.Context, cmd goredis.Cmder) error
	Pipeline() goredis.Pipeliner
	Close() error
	PoolStats() *goredis.PoolStats
	Publish(ctx context.Context, channel string, message interface{}) *goredis.IntCmd
	Subscribe(ctx context.Context, channels ...string) *goredis.PubSub
	Command(ctx context.Context) *goredis.CommandsInfoCmd
}

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

const (
	pkgName = "grabredis"

	// redis commands
	redisEval    = "EVAL"
	redisEvalSha = "EVALSHA"

	// redis err response checks
	redisErrNoScript = "NOSCRIPT "

	metricError   = "error"
	metricElapsed = "elapsed"

	tagCmdPrefix             = "grab_redis_cmd:"
	tagHostPrefix            = "grab_redis_host:"
	tagFunctionDo            = "grab_redis_func:do"
	tagFunctionPipeline      = "grab_redis_func:pipeline"
	tagFunctionRun           = "grab_redis_func:run"
	tagFunctionQueueLoadTest = "grab_redis_func:queueLoadTest"
	tagHystrixError          = "grab_redis_func:hystrix_error"
	tagHystrixTimeout        = "grab_redis_func:hystrix_timeout"
	tagHystrixCircuitOpen    = "grab_redis_func:hystrix_circuit_open"
	tagHystrixMaxConcurrency = "grab_redis_func:hystrix_max_concurrency"
	metricShutdown           = "shutdown"
	metricActive             = "active"
	metricTotal              = "total"

	tagTimeoutTrue  = "timeout:true"
	tagTimeoutFalse = "timeout:false"

	ucmEmptyString = "<nil>"

	// default config
	defaultHostAndPort     = "localhost:6379"
	defaultReadMode        = ModeReadFromSlaves
	defaultDialTimeoutInMs = 5000

	// default timeout
	defaultShutdownTimeout = 5 * time.Second
	reportInterval         = 5 * time.Second

	// default CB config
	defaultCBTimeoutInMS   = 31000
	defaultCBMaxConcurrent = 5000
	defaultCBErrPercent    = 50

	// load test scheduler
	defaultMaxChanSize       = 10000
	defaultMaxWorker         = 10
	defaultWorkerIdleTimeout = 1000
)

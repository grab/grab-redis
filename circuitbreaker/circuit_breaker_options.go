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

import gredis "github.com/grab/grab-redis"

// Option is a function that configures circuit breaker behaviors
type Option func(*cbOption)

// CbOption configures circuit breaker's optional behaviors
type cbOption struct {
	userErrHandler  IsNonThreatErr
	fallbackHandler func(error) error // Not using FallbackFunc here to avoid type conflict with hystrix.fallbackfn
	tags            []string
	unlockFeature   bool // Enables experimental features and directs traffic to beta version of cb
	logger          gredis.Logger
}

func newCBOption() *cbOption {
	return &cbOption{}
}

func makeOption(opts ...Option) *cbOption {
	options := newCBOption()
	for _, opt := range opts {
		opt(options)
	}

	// set default logger (if none was defined)
	if options.logger == nil {
		options.logger = gredis.NewNoopLogger()
	}

	// set user handler (if none was defined)
	if options.userErrHandler == nil {
		options.userErrHandler = getDefaultUserErrorHandler()
	}
	return options
}

// IsNonThreatErr is a function that checks the supplied error and decides if the circuit breaker should track the error
// (an threat error), or if it's an non-threat error that it should not track. This function returns 2 parameter: if
// this error is non-threat, and the error that circuit breaker should track. Three common use cases will be:
//
// 1. The error is not a threat error, e.g. "data not found" returned from mySQL driver:
// func (error) (bool, error) { return true, err }
//
// 2. The error is a threat error, and circuit breaker should track this error, e.g. status code 500:
// func (err error) (bool, error) { return false, err }
//
// 3. The error is a threat error, and circuit breaker should track a new error:
// func (err error) (bool, error) { return false, errors.New("Internal logic went wrong: " + err.Error()) }
//
// When nil error is passed, no error will be tracked
type IsNonThreatErr func(error) (nonThreat bool, errOut error)

// FallbackFunc is a function that will be called when error happens in cb function
type FallbackFunc func(error) error

// WithUserErrorHandler configures userErrorHandler. If error is safe, userErrorHandler should return
// true and nil; otherwise return false with an error.
func WithUserErrorHandler(handler IsNonThreatErr) Option {
	return func(opt *cbOption) {
		opt.userErrHandler = func(err error) (nonThreat bool, errOut error) {
			if err == nil {
				// nonThreat is not meaningful when errOut is nil, so return default value
				return false, nil
			}
			return handler(err)
		}
	}
}

// return the default user error handler
func getDefaultUserErrorHandler() func(err error) (nonThreat bool, errOut error) {
	return func(err error) (bool, error) {
		return false, err
	}
}

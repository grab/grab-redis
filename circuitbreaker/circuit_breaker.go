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
	"context"
	"github.com/myteksi/hystrix-go/hystrix"
	"github.com/myteksi/hystrix-go/hystrix/commandbuilder"
)

// CircuitBuilder wrapper over upstream/commandbuilder
type CircuitBuilder struct {
	*commandbuilder.CommandBuilder
}

// NewCircuitBuilder create a wrapper for commanduilder
func NewCircuitBuilder(circuitName string) *CircuitBuilder {
	return &CircuitBuilder{
		commandbuilder.New(circuitName),
	}
}

// AllowRequest returns true if the circuit is closed. When the circuit is open, this call will occasionally return true
// in order to allow for testing the health of a circuit
func AllowRequest(circuitName string) bool {
	circuit, _, err := hystrix.GetCircuit(circuitName)
	if err != nil {
		return true
	}
	return circuit.AllowRequest()
}

// IsCircuitOpen returns true if the circuit is open. Can be used before command execution to check whether it should be
// attempted or not.
func IsCircuitOpen(circuitName string) bool {
	circuit, _, err := hystrix.GetCircuit(circuitName)
	if err != nil {
		return false
	}
	return circuit.IsOpen()
}

func aggregateError(ctx context.Context, instrumentCtxErr func(error), softErrorCh, fatalErrorCh chan error) chan error {
	aggregatedErrors := make(chan error, 1)
	// note: we want to call Finish() on the span as soon as the error is available, *before* we send the error to the
	// channel, because there's no guarantee that the caller will ever read from the channel.
	go func() {
		select {
		case softError := <-softErrorCh:
			aggregatedErrors <- softError

		case fatalError := <-fatalErrorCh:
			aggregatedErrors <- fatalError

		case <-ctx.Done():
			ctxErr := ctx.Err()
			instrumentCtxErr(ctxErr)
			aggregatedErrors <- ctxErr
		}
	}()

	return aggregatedErrors
}

// ConfigureCircuit configs circuit breaker behavior for a circuit. It should be called before calling any `Go()` and `Do()`
// of this circuit.
func ConfigureCircuit(circuit *hystrix.Settings) {
	hystrix.Initialize(circuit)
}

// Go executes an action with protection of circuit breaker in async manner. It returns an error channel that provides
// the result of the routine. All possible errors are:
// --------soft errors------------
// 1. nil, if nothing goes wrong;
// 2. context error;
// 3. routine non-threat error;
// 4. routine panics error;
// --------fatal errors-----------
// 5. circuit breaker error: max concurrency reached;
// 6. circuit breaker error: routine timeout;
// 7. circuit breaker error: routine threat error.
//
// Please be reminded that threat error returned by routine would be wrapped up as circuit breaker error.
func Go(ctx context.Context, circuitName string, routine func() error, opts ...Option) chan error {
	option := makeOption(opts...)
	option.tags = append(option.tags, circuitName)

	// create a buffered channel in case there's no listener
	softErrorCh := make(chan error, 1)

	run := func() (output error) {
		// this error is the error return by routine, userErrorHandler should determine if this error
		// needs to trigger a circuit open.
		handleError := func(routineError error) error {
			if routineError != nil {
				isErrorSafe, err := option.userErrHandler(routineError)
				if !isErrorSafe {
					option.logger.ServiceDown(circuitName, err)
					return err
				}
			}

			softErrorCh <- routineError
			return nil
		}

		defer func() {
			if r := recover(); r != nil {
				err := RoutinePanicError{Recover: r}
				output = handleError(err)
			}
		}()

		err := routine()
		output = handleError(err)
		return output
	}

	// fallback handles is triggered when:
	// 1. circuit is already open;
	// 2. concurrency limit reach;
	// 3. routine timeout;
	// 4. unhealthy routine response.
	var fallback func(error) error
	if option.fallbackHandler != nil {
		// The ugly logic here is to prevent goroutine leak when some users install fallback as
		// func(err error) error { return nil }
		fallback = func(err error) error {
			errFallback := option.fallbackHandler(err)
			if errFallback == nil {
				softErrorCh <- nil
			}
			return errFallback
		}
	}
	instrumentCtxErr := func(err error) {
		if err != nil {
			option.logger.ContextError(circuitName, err)
		}
	}
	var fatalErrorCh chan error

	fatalErrorCh = hystrix.Go(circuitName, run, fallback)

	return aggregateError(ctx, instrumentCtxErr, softErrorCh, fatalErrorCh)
}

// Do execute an action with protection of circuit breaker in sync manner.
func Do(parent context.Context, name string, routine func() error, opts ...Option) error {
	return <-Go(parent, name, routine, opts...)
}

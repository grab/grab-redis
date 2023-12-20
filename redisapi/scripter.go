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
	"crypto/sha1"
	"encoding/hex"
	"io"
)

// Script encapsulates the source, hash and key count for a Lua script. See
// http://redis.io/commands/eval for information on scripts in Redis.
type Script struct {
	keyCount int
	src      string
	hash     string
}

// NewScript returns a new script object. If keyCount is greater than or equal to zero, then the count is
// automatically inserted in the EVAL command argument list.
// If keyCount is less than zero, then the application supplies the count as the first value in the
// keysAndArgs argument as the first argument in the call to Redis.
func NewScript(keyCount int, scriptSource string) *Script {
	h := sha1.New()
	_, _ = io.WriteString(h, scriptSource)
	return &Script{keyCount, scriptSource, hex.EncodeToString(h.Sum(nil))}
}

func (s *Script) args(spec string, keysAndArgs []interface{}) []interface{} {
	var args []interface{}
	if s.keyCount < 0 {
		args = make([]interface{}, 1+len(keysAndArgs))
		args[0] = spec
		copy(args[1:], keysAndArgs)
	} else {
		args = make([]interface{}, 2+len(keysAndArgs))
		args[0] = spec
		args[1] = s.keyCount
		copy(args[2:], keysAndArgs)
	}
	return args
}

// GetHashAndArgs will return the args for running the script (via EVALSHA)
func (s *Script) GetHashAndArgs(keysAndArgs ...interface{}) []interface{} {
	return s.args(s.hash, keysAndArgs)
}

// GetScriptAndArgs will return the args for running the script (via EVAL)
func (s *Script) GetScriptAndArgs(keysAndArgs ...interface{}) []interface{} {
	return s.args(s.src, keysAndArgs)
}

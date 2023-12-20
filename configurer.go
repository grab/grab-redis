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

import "fmt"

// Configurer is an interface that allows the client to be configured dynamically, service teams can implement this interface to provide their own configuration.
type Configurer interface {
	OnChange(callback func() error)
	Unmarshal(obj interface{}) error
}

type staticConfigurer struct {
	*ConnectorConfig
}

func newStaticConfigurer(config *ConnectorConfig) *staticConfigurer {
	return &staticConfigurer{
		config,
	}
}

func (s *staticConfigurer) OnChange(func() error) {}

func (s *staticConfigurer) Unmarshal(obj interface{}) error {
	config, ok := obj.(*ConnectorConfig)
	if !ok {
		return fmt.Errorf("staticConfigurer Unmarshalling failed")
	}
	*config = *s.ConnectorConfig
	return nil
}

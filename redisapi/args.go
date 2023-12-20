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

import "reflect"

// Args is a helper for constructing command arguments from structured values.
type Args []interface{}

// Add returns the result of appending value to args.
func (args *Args) Add(values ...interface{}) *Args {
	*args = append(*args, values...)
	return args
}

// AddFlat returns the result of appending the flattened value of v to args.
//
// Maps are flattened by appending the alternating keys and map values to args.
//
// Slices are flattened by appending the slice elements to args.
//
// Structs are flattened by appending the alternating names and values of
// exported fields to args. If v is a nil struct pointer, then nothing is
// appended. The 'redis' field tag overrides struct field names. See
// redigo.ScanStruct for more information on the use of the 'redis' field tag.
//
// Other types are appended to args as is.
func (args *Args) AddFlat(v interface{}) *Args {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Struct:
		*args = flattenStruct(*args, rv)
	case reflect.Slice:
		for i := 0; i < rv.Len(); i++ {
			*args = append(*args, rv.Index(i).Interface())
		}
	case reflect.Map:
		for _, k := range rv.MapKeys() {
			*args = append(*args, k.Interface(), rv.MapIndex(k).Interface())
		}
	case reflect.Ptr:
		if rv.Type().Elem().Kind() == reflect.Struct {
			if !rv.IsNil() {
				*args = flattenStruct(*args, rv.Elem())
			}
		} else {
			*args = append(*args, v)
		}
	default:
		*args = append(*args, v)
	}
	return args
}

// NewArgs makes a new *Args
func NewArgs(args ...interface{}) *Args {
	result := &Args{}
	*result = Args(args)
	return result
}

// Value return Args type the pointer points to
func (args *Args) Value() Args {
	return *args
}

func flattenStruct(args Args, v reflect.Value) Args {
	ss := structSpecForType(v.Type())
	for _, fs := range ss.l {
		fv := v.FieldByIndex(fs.index)
		if fs.omitEmpty {
			var empty = false
			switch fv.Kind() {
			case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
				empty = fv.Len() == 0
			case reflect.Bool:
				empty = !fv.Bool()
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				empty = fv.Int() == 0
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
				empty = fv.Uint() == 0
			case reflect.Float32, reflect.Float64:
				empty = fv.Float() == 0
			case reflect.Interface, reflect.Ptr:
				empty = fv.IsNil()
			}
			if empty {
				continue
			}
		}
		args = append(args, fs.name, fv.Interface())
	}
	return args
}

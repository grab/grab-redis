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
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

func ensureLen(d reflect.Value, n int) {
	if n > d.Cap() {
		d.Set(reflect.MakeSlice(d.Type(), n, n))
	} else {
		d.SetLen(n)
	}
}

func cannotConvert(d reflect.Value, s interface{}) error {
	var sname string
	switch s.(type) {
	case string:
		sname = "Redis simple string"
	case int64:
		sname = "Redis integer"
	case []byte:
		sname = "Redis bulk string"
	case []interface{}:
		sname = "Redis array"
	default:
		sname = reflect.TypeOf(s).String()
	}
	return fmt.Errorf("cannot convert from %s to %s", sname, d.Type())
}

func convertAssignBulkString(d reflect.Value, s []byte) (err error) {
	switch d.Type().Kind() {
	case reflect.Float32, reflect.Float64:
		var x float64
		x, err = strconv.ParseFloat(string(s), d.Type().Bits())
		d.SetFloat(x)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var x int64
		x, err = strconv.ParseInt(string(s), 10, d.Type().Bits())
		d.SetInt(x)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var x uint64
		x, err = strconv.ParseUint(string(s), 10, d.Type().Bits())
		d.SetUint(x)
	case reflect.Bool:
		var x bool
		x, err = strconv.ParseBool(string(s))
		d.SetBool(x)
	case reflect.String:
		d.SetString(string(s))
	case reflect.Slice:
		if d.Type().Elem().Kind() != reflect.Uint8 {
			err = cannotConvert(d, s)
		} else {
			d.SetBytes(s)
		}
	default:
		err = cannotConvert(d, s)
	}
	return
}

func convertAssignSimpleString(d reflect.Value, s string) (err error) {
	switch d.Type().Kind() {
	case reflect.Float32, reflect.Float64:
		var x float64
		x, err = strconv.ParseFloat(s, d.Type().Bits())
		d.SetFloat(x)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var x int64
		x, err = strconv.ParseInt(s, 10, d.Type().Bits())
		d.SetInt(x)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var x uint64
		x, err = strconv.ParseUint(s, 10, d.Type().Bits())
		d.SetUint(x)
	case reflect.Bool:
		var x bool
		x, err = strconv.ParseBool(s)
		d.SetBool(x)
	case reflect.String:
		d.SetString(s)
	case reflect.Slice:
		if d.Type().Elem().Kind() != reflect.Uint8 {
			err = cannotConvert(d, s)
		} else {
			d.SetBytes([]byte(s))
		}
	default:
		err = cannotConvert(d, s)
	}
	return
}

func convertAssignInt(d reflect.Value, s int64) (err error) {
	switch d.Type().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		d.SetInt(s)
		if d.Int() != s {
			err = strconv.ErrRange
			d.SetInt(0)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if s < 0 {
			err = strconv.ErrRange
		} else {
			x := uint64(s)
			d.SetUint(x)
			if d.Uint() != x {
				err = strconv.ErrRange
				d.SetUint(0)
			}
		}
	case reflect.Bool:
		d.SetBool(s != 0)
	default:
		err = cannotConvert(d, s)
	}
	return
}

func convertAssignValue(d reflect.Value, s interface{}) (err error) {
	switch s := s.(type) {
	case []byte:
		err = convertAssignBulkString(d, s)
	case string:
		err = convertAssignSimpleString(d, s)
	case int64:
		err = convertAssignInt(d, s)
	default:
		err = cannotConvert(d, s)
	}
	return err
}

func convertAssignArray(d reflect.Value, s []interface{}) error {
	if d.Type().Kind() != reflect.Slice {
		return cannotConvert(d, s)
	}
	ensureLen(d, len(s))
	for i := 0; i < len(s); i++ {
		if err := convertAssignValue(d.Index(i), s[i]); err != nil {
			return err
		}
	}
	return nil
}

func convertAssign(d interface{}, s interface{}) (err error) {
	// Handle the most common destination types using type switches and
	// fall back to reflection for all other types.
	switch s := s.(type) {
	case nil:
		// ignore
	case []byte:
		switch d := d.(type) {
		case *string:
			*d = string(s)
		case *int:
			*d, err = strconv.Atoi(string(s))
		case *bool:
			*d, err = strconv.ParseBool(string(s))
		case *[]byte:
			*d = s
		case *interface{}:
			*d = s
		case nil:
			// skip value
		default:
			if d := reflect.ValueOf(d); d.Type().Kind() != reflect.Ptr {
				err = cannotConvert(d, s)
			} else {
				err = convertAssignBulkString(d.Elem(), s)
			}
		}
	case int64:
		switch d := d.(type) {
		case *int:
			x := int(s)
			if int64(x) != s {
				err = strconv.ErrRange
				x = 0
			}
			*d = x
		case *bool:
			*d = s != 0
		case *interface{}:
			*d = s
		case nil:
			// skip value
		default:
			if d := reflect.ValueOf(d); d.Type().Kind() != reflect.Ptr {
				err = cannotConvert(d, s)
			} else {
				err = convertAssignInt(d.Elem(), s)
			}
		}
	case string:
		switch d := d.(type) {
		case *string:
			*d = s
		default:
			if d := reflect.ValueOf(d); d.Type().Kind() != reflect.Ptr {
				err = cannotConvert(d, s)
			} else {
				err = convertAssignSimpleString(d.Elem(), s)
			}
		}
	case []interface{}:
		switch d := d.(type) {
		case *[]interface{}:
			*d = s
		case *interface{}:
			*d = s
		case nil:
			// skip value
		default:
			if d := reflect.ValueOf(d); d.Type().Kind() != reflect.Ptr {
				err = cannotConvert(d, s)
			} else {
				err = convertAssignArray(d.Elem(), s)
			}
		}
	default:
		err = cannotConvert(reflect.ValueOf(d), s)
	}
	return
}

// Scan copies from src to the values pointed at by dest.
//
// The values pointed at by dest must be an integer, float, boolean, string,
// []byte, interface{} or slices of these types. Scan uses the standard strconv
// package to convert bulk strings to numeric and boolean types.
//
// If a dest value is nil, then the corresponding src value is skipped.
//
// If a src element is nil, then the corresponding dest value is not modified.
//
// To enable easy use of Scan in a loop, Scan returns the slice of src
// following the copied values.
func Scan(src []interface{}, dest ...interface{}) ([]interface{}, error) {

	if len(src) < len(dest) {
		return nil, errors.New("redis.Scan: array short")
	}
	var err error
	for i, d := range dest {
		err = convertAssign(d, src[i])
		if err != nil {
			err = fmt.Errorf("redis.Scan: cannot assign to dest %d: %v", i, err)
			break
		}
	}
	return src[len(dest):], err
}

type fieldSpec struct {
	name      string
	index     []int
	omitEmpty bool
}

type structSpec struct {
	m map[string]*fieldSpec
	l []*fieldSpec
}

func (ss *structSpec) fieldSpec(name []byte) *fieldSpec {
	return ss.m[string(name)]
}

func compileStructSpec(t reflect.Type, depth map[string]int, index []int, ss *structSpec) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		switch {
		case f.PkgPath != "" && !f.Anonymous:
			// Ignore unexported fields.
		case f.Anonymous:
			// TODO: Handle pointers. Requires change to decoder and
			// protection against infinite recursion.
			if f.Type.Kind() == reflect.Struct {
				compileStructSpec(f.Type, depth, append(index, i), ss)
			}
		default:
			fs := &fieldSpec{name: f.Name}
			tag := f.Tag.Get("redis")
			p := strings.Split(tag, ",")
			if len(p) > 0 {
				if p[0] == "-" {
					continue
				}
				if len(p[0]) > 0 {
					fs.name = p[0]
				}
				for _, s := range p[1:] {
					switch s {
					case "omitempty":
						fs.omitEmpty = true
					default:
						panic(fmt.Errorf("redis: unknown field tag %s for type %s", s, t.Name()))
					}
				}
			}
			d, found := depth[fs.name]
			if !found {
				d = 1 << 30
			}
			switch {
			case len(index) == d:
				// At same depth, remove from result.
				delete(ss.m, fs.name)
				j := 0
				for ii := 0; ii < len(ss.l); ii++ {
					if fs.name != ss.l[ii].name {
						ss.l[j] = ss.l[ii]
						j++
					}
				}
				ss.l = ss.l[:j]
			case len(index) < d:
				fs.index = make([]int, len(index)+1)
				copy(fs.index, index)
				fs.index[len(index)] = i
				depth[fs.name] = len(index)
				ss.m[fs.name] = fs
				ss.l = append(ss.l, fs)
			}
		}
	}
}

var (
	structSpecMutex sync.RWMutex
	structSpecCache = make(map[reflect.Type]*structSpec)
)

func structSpecForType(t reflect.Type) *structSpec {

	structSpecMutex.RLock()
	ss, found := structSpecCache[t]
	structSpecMutex.RUnlock()
	if found {
		return ss
	}

	structSpecMutex.Lock()
	defer structSpecMutex.Unlock()
	ss, found = structSpecCache[t]
	if found {
		return ss
	}

	ss = &structSpec{m: make(map[string]*fieldSpec)}
	compileStructSpec(t, make(map[string]int), nil, ss)
	structSpecCache[t] = ss
	return ss
}

var errScanStructValue = errors.New("redis.ScanStruct: value must be non-nil pointer to a struct")

// ScanStruct scans alternating names and values from src to a struct. The
// HGETALL and CONFIG GET commands return replies in this format.
//
// ScanStruct uses exported field names to match values in the response. Use
// 'redis' field tag to override the name:
//
//      Field int `redis:"myName"`
//
// Fields with the tag redis:"-" are ignored.
//
// Integer, float, boolean, string and []byte fields are supported. Scan uses the
// standard strconv package to convert bulk string values to numeric and
// boolean types.
//
// If a src element is nil, then the corresponding field is not modified.
func ScanStruct(src []interface{}, dest interface{}) error {
	d := reflect.ValueOf(dest)
	if d.Kind() != reflect.Ptr || d.IsNil() {
		return errScanStructValue
	}
	d = d.Elem()
	if d.Kind() != reflect.Struct {
		return errScanStructValue
	}
	ss := structSpecForType(d.Type())

	if len(src)%2 != 0 {
		return errors.New("redis.ScanStruct: number of values not a multiple of 2")
	}

	for i := 0; i < len(src); i += 2 {
		s := src[i+1]
		if s == nil {
			continue
		}
		name, ok := src[i].([]byte)
		if !ok {
			return fmt.Errorf("redis.ScanStruct: key %d not a bulk string value", i)
		}
		fs := ss.fieldSpec(name)
		if fs == nil {
			continue
		}
		if err := convertAssignValue(d.FieldByIndex(fs.index), s); err != nil {
			return fmt.Errorf("redis.ScanStruct: cannot assign field %s: %v", fs.name, err)
		}
	}
	return nil
}

var (
	errScanSliceValue = errors.New("redis.ScanSlice: dest must be non-nil pointer to a struct")
)

// ScanSlice scans src to the slice pointed to by dest. The elements the dest
// slice must be integer, float, boolean, string, struct or pointer to struct
// values.
//
// Struct fields must be integer, float, boolean or string values. All struct
// fields are used unless a subset is specified using fieldNames.
func ScanSlice(src []interface{}, dest interface{}, fieldNames ...string) error {
	d := reflect.ValueOf(dest)
	if d.Kind() != reflect.Ptr || d.IsNil() {
		return errScanSliceValue
	}
	d = d.Elem()
	if d.Kind() != reflect.Slice {
		return errScanSliceValue
	}

	isPtr := false
	t := d.Type().Elem()
	if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct {
		isPtr = true
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		ensureLen(d, len(src))
		for i, s := range src {
			if s == nil {
				continue
			}
			if err := convertAssignValue(d.Index(i), s); err != nil {
				return fmt.Errorf("redis.ScanSlice: cannot assign element %d: %v", i, err)
			}
		}
		return nil
	}

	ss := structSpecForType(t)
	fss := ss.l
	if len(fieldNames) > 0 {
		fss = make([]*fieldSpec, len(fieldNames))
		for i, name := range fieldNames {
			fss[i] = ss.m[name]
			if fss[i] == nil {
				return fmt.Errorf("redis.ScanSlice: ScanSlice bad field name %s", name)
			}
		}
	}

	if len(fss) == 0 {
		return errors.New("redis.ScanSlice: no struct fields")
	}

	n := len(src) / len(fss)
	if n*len(fss) != len(src) {
		return errors.New("redis.ScanSlice: length not a multiple of struct field count")
	}

	ensureLen(d, n)
	for i := 0; i < n; i++ {
		di := d.Index(i)
		if isPtr {
			if di.IsNil() {
				di.Set(reflect.New(t))
			}
			di = di.Elem()
		}
		for j, fs := range fss {
			s := src[i*len(fss)+j]
			if s == nil {
				continue
			}
			if err := convertAssignValue(di.FieldByIndex(fs.index), s); err != nil {
				return fmt.Errorf("redis.ScanSlice: cannot assign element %d to field %s: %v", i*len(fss)+j, fs.name, err)
			}
		}
	}
	return nil
}

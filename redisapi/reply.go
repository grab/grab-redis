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
	"strconv"
)

var (
	// ErrNoData indicates that a reply value is nil.
	ErrNoData = errors.New("redis: nil returned")
)

// Int is a helper that converts a command reply to an integer.
func Int(reply interface{}, err error) (int, error) {
	if err != nil {
		return 0, err
	}
	switch reply := reply.(type) {
	case int64:
		x := int(reply)
		if int64(x) != reply {
			return 0, strconv.ErrRange
		}
		return x, nil
	case []byte:
		n, err := strconv.ParseInt(string(reply), 10, 0)
		return int(n), err
	case string:
		n, err := strconv.ParseInt(reply, 10, 0)
		return int(n), err
	case nil:
		return 0, ErrNoData
	default:
		return 0, fmt.Errorf("redis: unexpected type for Int, got type %T", reply)
	}
}

// Int64 is a helper that converts a command reply to 64 bit integer.
func Int64(reply interface{}, err error) (int64, error) {
	if err != nil {
		return 0, err
	}
	switch reply := reply.(type) {
	case int64:
		return reply, nil
	case []byte:
		n, err := strconv.ParseInt(string(reply), 10, 64)
		return n, err
	case string:
		n, err := strconv.ParseInt(reply, 10, 64)
		return n, err
	case nil:
		return 0, ErrNoData
	default:
		return 0, fmt.Errorf("redis: unexpected type for Int64, got type %T", reply)
	}
}

// Uint64 is a helper that converts a command reply to 64 bit unsigned integer.
func Uint64(reply interface{}, err error) (uint64, error) {
	if err != nil {
		return 0, err
	}
	switch reply := reply.(type) {
	case int64:
		if reply < 0 {
			return 0, fmt.Errorf("redis: unexpected value for Uint64, got %d", reply)
		}
		return uint64(reply), nil
	case []byte:
		n, err := strconv.ParseUint(string(reply), 10, 64)
		return n, err
	case string:
		n, err := strconv.ParseUint(reply, 10, 64)
		return n, err
	case nil:
		return 0, ErrNoData
	default:
		return 0, fmt.Errorf("redis: unexpected type for Uint64, got type %T", reply)
	}
}

// Float64 is a helper that converts a command reply to 64 bit float.
func Float64(reply interface{}, err error) (float64, error) {
	if err != nil {
		return 0, err
	}
	switch reply := reply.(type) {
	case []byte:
		n, err := strconv.ParseFloat(string(reply), 64)
		return n, err
	case string:
		n, err := strconv.ParseFloat(reply, 64)
		return n, err
	case nil:
		return 0, ErrNoData
	default:
		return 0, fmt.Errorf("redis: unexpected type for Float64, got type %T", reply)
	}
}

// String is a helper that converts a command reply to a string.
func String(reply interface{}, err error) (string, error) {
	if err != nil {
		return "", err
	}
	switch reply := reply.(type) {
	case []byte:
		return string(reply), nil
	case string:
		return reply, nil
	case nil:
		return "", ErrNoData
	default:
		return "", fmt.Errorf("redis: unexpected type for String, got type %T", reply)
	}
}

// Bytes is a helper that converts a command reply to a slice of bytes.
func Bytes(reply interface{}, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}

	switch b := reply.(type) {
	case []byte:
		return b, nil
	case string:
		return []byte(b), nil
	case nil:
		return nil, ErrNoData
	default:
		return nil, fmt.Errorf("redis: unexpected type=%T for []byte", reply)
	}
}

// Bool is a helper that converts a command reply to a boolean.
func Bool(reply interface{}, err error) (bool, error) {
	if err != nil {
		return false, err
	}
	switch reply := reply.(type) {
	case int64:
		return reply != 0, nil
	case []byte:
		return strconv.ParseBool(string(reply))
	case string:
		return strconv.ParseBool(reply)
	case nil:
		return false, ErrNoData
	default:
		return false, fmt.Errorf("redis: unexpected type for Bool, got type %T", reply)
	}
}

// Values is a helper that converts an array command reply to a []interface{}.
func Values(reply interface{}, err error) ([]interface{}, error) {
	if err != nil {
		return nil, err
	}
	switch reply := reply.(type) {
	case []interface{}:
		return reply, nil
	case nil:
		return nil, ErrNoData
	default:
		return nil, fmt.Errorf("redis: unexpected type for Values, got type %T", reply)
	}
}

// Strings is a helper that converts an array command reply to a []string.
func Strings(reply interface{}, err error) ([]string, error) {
	if err != nil {
		return nil, err
	}
	switch reply := reply.(type) {
	case []interface{}:
		result := make([]string, len(reply))
		for i := range reply {
			if reply[i] == nil {
				continue
			}
			var ok bool
			result[i], ok = toString(reply[i])
			if !ok {
				return nil, fmt.Errorf("redis: unexpected element type for Strings, got type %T", reply[i])
			}
		}
		return result, nil
	case nil:
		return nil, ErrNoData
	default:
		return nil, fmt.Errorf("redis: unexpected type for Strings, got type %T", reply)
	}
}

func toString(i interface{}) (string, bool) {
	switch s := i.(type) {
	case []byte:
		return string(s), true
	case string:
		return s, true
	default:
		return "", false
	}
}

// ByteSlices is a helper that converts an array command reply to a [][]byte.
func ByteSlices(reply interface{}, err error) ([][]byte, error) {
	values, err := Values(reply, err)
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, nil
	}

	slices := make([][]byte, len(values))
	for i := range slices {
		if values[i] == nil {
			continue
		}

		slices[i], err = Bytes(values[i], nil)
		if err != nil {
			return nil, err
		}
	}

	return slices, nil
}

// Ints is a helper that converts an array command reply to a []int.
func Ints(reply interface{}, err error) ([]int, error) {
	var ints []int
	values, err := Values(reply, err)
	if err != nil {
		return ints, err
	}
	if err := ScanSlice(values, &ints); err != nil {
		return ints, err
	}
	return ints, nil
}

// StringMap is a helper that converts an array of strings (alternating key, value)
// into a map[string]string. The HGETALL and CONFIG GET commands return replies in this format.
// Requires an even number of values in reply.
func StringMap(reply interface{}, err error) (map[string]string, error) {
	values, err := Values(reply, err)
	if err != nil {
		return nil, err
	}
	if len(values)%2 != 0 {
		return nil, errors.New("redis: StringMap expects even number of values result")
	}
	m := make(map[string]string, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, err := String(values[i], nil)
		if err != nil {
			return nil, errors.New("redis: ScanMap key not a bulk string value")
		}
		value, err := String(values[i+1], nil)
		if err != nil {
			return nil, errors.New("redis: ScanMap value not a bulk string value")
		}
		m[key] = value
	}
	return m, nil
}

// IntMap is a helper that converts an array of strings (alternating key, value)
// into a map[string]int. The HGETALL commands return replies in this format.
// Requires an even number of values in result.
func IntMap(result interface{}, err error) (map[string]int, error) {
	values, err := Values(result, err)
	if err != nil {
		return nil, err
	}
	if len(values)%2 != 0 {
		return nil, errors.New("redis: IntMap expects even number of values result")
	}
	m := make(map[string]int, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, err := String(values[i], nil)
		if err != nil {
			return nil, errors.New("redis: ScanMap key not a bulk string value")
		}
		value, err := Int(values[i+1], nil)
		if err != nil {
			return nil, errors.New("redis: ScanMap value not a bulk int value")
		}
		m[key] = value
	}
	return m, nil
}

// Int64Map is a helper that converts an array of strings (alternating key, value)
// into a map[string]int64. The HGETALL commands return replies in this format.
// Requires an even number of values in result.
func Int64Map(result interface{}, err error) (map[string]int64, error) {
	values, err := Values(result, err)
	if err != nil {
		return nil, err
	}
	if len(values)%2 != 0 {
		return nil, errors.New("redis: Int64Map expects even number of values result")
	}
	m := make(map[string]int64, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, err := String(values[i], nil)
		if err != nil {
			return nil, errors.New("redis: ScanMap key not a bulk string value")
		}
		value, err := Int64(values[i+1], nil)
		if err != nil {
			return nil, errors.New("redis: ScanMap value not a bulk int64 value")
		}
		m[key] = value
	}
	return m, nil
}

// Float64Map is a helper that converts an array of strings (alternating key, value)
// into a map[string]float64. The HGETALL commands return replies in this format.
// Requires an even number of values in result.
func Float64Map(result interface{}, err error) (map[string]float64, error) {
	values, err := Values(result, err)
	if err != nil {
		return nil, err
	}
	if len(values)%2 != 0 {
		return nil, errors.New("redis: Float64Map expects even number of values result")
	}
	m := make(map[string]float64, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, err := String(values[i], nil)
		if err != nil {
			return nil, errors.New("redis: ScanMap key not a bulk string value")
		}
		value, err := Float64(values[i+1], nil)
		if err != nil {
			return nil, errors.New("redis: ScanMap value not a bulk float64 value")
		}
		m[key] = value
	}
	return m, nil
}

// Float64s is a helper that converts an array command reply to a []Float64.
func Float64s(reply interface{}, err error) ([]float64, error) {
	values, err := Values(reply, err)
	if err != nil {
		return nil, err
	}
	floats := make([]float64, len(values))
	for i, value := range values {
		f, err := Float64(value, nil)
		if err != nil {
			return nil, err
		}
		floats[i] = f
	}
	return floats, nil
}

// Int64s is a helper that converts an array command reply to a []int64.
func Int64s(reply interface{}, err error) ([]int64, error) {
	var ints []int64
	values, err := Values(reply, err)
	if err != nil {
		return nil, err
	}
	if err := ScanSlice(values, &ints); err != nil {
		return nil, err
	}
	return ints, nil
}

// Int64StringMapSlice is a helper that converts the reply of a ZRANGE ... WITHSCORES (alternating key, value)
// into a map[int64][]string.
func Int64StringMapSlice(reply interface{}, err error) (map[int64][]string, error) {
	if err != nil {
		return nil, err
	}

	values, _ := reply.([]interface{})
	if len(values)%2 != 0 {
		return nil, errors.New("expects even number of values")
	}

	out := make(map[int64][]string, len(values)/2)
	for index := 0; index < len(values); index += 2 {
		value, err := String(values[index], nil)
		if err != nil {
			return nil, errors.New("redis: ScanMap key not a bulk string value")
		}

		score, err := Int64(values[index+1], nil)
		if err != nil {
			return nil, errors.New("redis: ScanMap value not a bulk int64 value")
		}

		out[score] = append(out[score], value)
	}
	return out, nil
}

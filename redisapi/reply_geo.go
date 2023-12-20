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
)

// error constants
var (
	ErrConvertToGeoPointInvalidLength    = errors.New("failed to convert reply to GeoPosition: reply length is not 2")
	ErrConvertToGeoLocationInvalidLength = errors.New("failed to convert pReply to GeoPosition: pReply length must be 1-4")
)

// GeoPoint Represents a Physical GeoPoint in geographic notation [lat, lng].
type GeoPoint struct {
	lat float64
	lng float64
}

// NewGeoPoint Returns a new GeoPoint populated by the passed in latitude (lat) and longitude (lng) values.
func NewGeoPoint(lat float64, lng float64) GeoPoint {
	return GeoPoint{lat: lat, lng: lng}
}

// Lat Returns GeoPoint p's latitude.
func (p GeoPoint) Lat() float64 {
	return p.lat
}

// Lng Returns GeoPoint p's longitude.
func (p GeoPoint) Lng() float64 {
	return p.lng
}

// GeoLocation is a struct which represents the response of some geo related commands
type GeoLocation struct {
	Name     string
	Distance float64
	GeoPoint GeoPoint
	Hash     int64
}

// GeoPosition converts a reply to the GeoPoint
// For the response of the below commands
// https://redis.io/commands/geopos
func GeoPosition(reply interface{}, err error) ([]GeoPoint, error) {
	values, err := Values(reply, err)
	if err != nil {
		return nil, err
	}
	geoPoints := make([]GeoPoint, len(values))
	for i, value := range values {
		geoPoints[i], err = convertToGeoGeoPoint(value)
		if err != nil {
			return nil, err
		}
	}
	return geoPoints, nil
}

// GeoLocations converts a reply to the []*GeoLocation
// For the response of the below commands
// https://redis.io/commands/georadius
// https://redis.io/commands/georadiusbymember
func GeoLocations(reply interface{}, err error) ([]*GeoLocation, error) {
	values, err := Values(reply, err)
	if err != nil {
		return nil, err
	}
	list := make([]*GeoLocation, len(values))
	for i, v := range values {
		list[i], err = convertToGeoLocation(v)
		if err != nil {
			return nil, err
		}
	}
	return list, nil
}

// convertToGeoGeoPoint converts interface{} to GeoPoint
func convertToGeoGeoPoint(reply interface{}) (GeoPoint, error) {
	latlng, err := Float64s(reply, nil)
	if err != nil {
		return GeoPoint{}, err
	}
	if len(latlng) != 2 {
		return GeoPoint{}, ErrConvertToGeoPointInvalidLength
	}
	return NewGeoPoint(latlng[1], latlng[0]), nil
}

// convertToGeoLocation converts []interface{} to a GeoLocation struct
func convertToGeoLocation(reply interface{}) (*GeoLocation, error) {
	result := &GeoLocation{}
	pReply, err := Values(reply, nil)
	if err != nil {
		return nil, err
	}
	pReplyLen := len(pReply)
	if pReplyLen == 0 || pReplyLen > 4 {
		return nil, ErrConvertToGeoLocationInvalidLength
	}
	for i, item := range pReply {
		if i == 0 {
			result.Name, err = String(item, nil)
			if err != nil {
				return nil, err
			}
			continue
		}
		switch item.(type) {
		case int64:
			result.Hash, err = Int64(item, nil)
			if err != nil {
				return nil, err
			}
		case string:
			result.Distance, err = Float64(item, nil)
			if err != nil {
				return nil, err
			}
		case []interface{}:
			result.GeoPoint, err = convertToGeoGeoPoint(item)
			if err != nil {
				return nil, err
			}
		}
	}
	return result, nil
}

// Copyright 2014 The zhgo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package web

import (
	"log"
	"reflect"
	"strconv"
)

// Parameter Conversion
func parameterConversion(str string, kind reflect.Kind) (reflect.Value, error) {
	var v reflect.Value
	var err error

	switch kind {
	case reflect.String:
		v = reflect.ValueOf(str)
	case reflect.Int64:
		var i64 int64
		i64, err = strconv.ParseInt(str, 10, 64)
		v = reflect.ValueOf(i64)
	default:
		log.Printf("String convert to int failure: %s\n", str)
	}

	return v, err
}

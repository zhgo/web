// Copyright 2014 The zhgo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package web

import (
	"net/http"
	"reflect"
)

// Action result
type Result struct {
	// Data
	Data interface{} `json:"data"`

	// error
	Err string `json:"err"`
}

// Registered controllers
var controllers map[string]map[string]reflect.Value = make(map[string]map[string]reflect.Value)

// Action load function
type ActionLoadFunc func(r *Request) Result

// App
var App Application

// *http.ServeMux
var muxList = make(map[string]*http.ServeMux)

// Alias of map[string]interface{}
type BodyJson map[string]interface{}

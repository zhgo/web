// Copyright 2014 The zhgo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package web

import (
	"reflect"
)

// Controller struct
type Controller struct {
	// Request
	Request *Request
}

// Failure
func (c *Controller) Fail(err error) ActionResult {
	return ActionResult{"", err.Error()}
}

// Success
func (c *Controller) Done(data interface{}) ActionResult {
	return ActionResult{data, ""}
}

// New router register
func NewController(module string, c interface{}) {
	if _, ok := controllers[module]; !ok {
		controllers[module] = make(map[string]reflect.Value)
	}

	value := reflect.ValueOf(c)
	controllers[module][value.Elem().Type().Name()] = value
}

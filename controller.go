// Copyright 2014 The zhgo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package web

import (
	"log"
	"reflect"
)

//Registered controllers
var controllers map[string]map[string]reflect.Value = make(map[string]map[string]reflect.Value)

//controller struct
type Controller struct {
	//Request
	Request *Request

	//Responsewriter
	View *View
}

//General error page
func (c *Controller) IndexErr(status int16, msg string) Result {
	log.Printf("SystemIndexErr\n")

	return c.View.Render()
}

//new router register
func NewController(module string, c interface{}) {
    if _, ok := controllers[module]; !ok {
        controllers[module] = make(map[string]reflect.Value)
    }

	value := reflect.ValueOf(c)
	controllers[module][value.Elem().Type().Name()] = value
}

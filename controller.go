// Copyright 2014 The zhgo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package web

import (
	"encoding/json"
	"github.com/zhgo/console"
	"github.com/zhgo/db"
	"reflect"
)

//Registered controllers
var controllers map[string]map[string]reflect.Value = make(map[string]map[string]reflect.Value)

//controller struct
type Controller struct {
	//Request
	Request *Request
}

// Add
func (c *Controller) Add() interface{} {
	var d map[string]interface{}
	err := json.NewDecoder(c.Request.HTTPRequest.Body).Decode(&d)
	if err != nil {
		return Result{-1, err.Error()}
	}

	if s, ok := db.Servers[App.Modules[c.Request.Module].DB.Name]; ok {
		tbName := console.CamelcaseToUnderscore(c.Request.Controller)
		r, err := s.InsertInto(tbName).Exec(d)
		if err != nil {
			return Result{-2, err.Error()}
		}

		return Result{r.LastInsertId, ""}
	} else {
		return Result{-1, "No database configuration"}
	}

	return d
}

// Detail
func (c *Controller) Detail(id string) interface{} {
	return map[string]string{"code": "1", "err": "Detail: Wrong paramaters"}
}

// Update
func (c *Controller) Edit(id string) interface{} {
	return map[string]string{"code": "1", "err": "Edit: Wrong paramaters"}
}

// Delete
func (c *Controller) Delete(id string) interface{} {
	return map[string]string{"code": "1", "err": "Delete: Wrong paramaters"}
}

// List
func (c *Controller) List(pi, ps int64) interface{} {
	return map[string]string{"code": "1", "err": "List: Wrong paramaters"}
}

//new router register
func NewController(module string, c interface{}) {
	if _, ok := controllers[module]; !ok {
		controllers[module] = make(map[string]reflect.Value)
	}

	value := reflect.ValueOf(c)
	controllers[module][value.Elem().Type().Name()] = value
}

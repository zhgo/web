// Copyright 2014 The zhgo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package web

import (
	"encoding/json"
	"fmt"
	"github.com/zhgo/dump"
	"github.com/zhgo/nameconv"
	"log"
	"net/http"
	"reflect"
	"strings"
)

// Controller interface
type Controller interface {
	Load(mdl, obj, act string, r *http.Request) error
	Render() Result
}

// Action result
type Result struct {
	// Data
	Data interface{} `json:"data"`

	// error
	Err string `json:"err"`
}

// Registe Controller
func NewController(c Controller) {
	ct := reflect.TypeOf(c).Elem()
	la := nameconv.CamelcaseToSlice(ct.Name(), false, -1)
	if len(la) < 3 {
		log.Fatalf("ControllerType.Name() parase error: %s\n", ct.Name())
	}

	l := len(la)
	mdl := la[0]
	obj := strings.Join(la[1:l-1], "")
	act := strings.Join(la[l-1:], "")
	pattern := fmt.Sprintf("/%s/%s/%s", mdl, obj, act)

	log.Printf("%#v\n", pattern)

	m, ok := app.Modules[mdl]
	if !ok {
		log.Fatalf("Module not found: %s\n", mdl)
	}

	if _, ok = app.muxList[m.Listen]; !ok {
		app.muxList[m.Listen] = http.NewServeMux()
	}

	app.muxList[m.Listen].HandleFunc(pattern, handle(mdl, obj, act, ct))
}

// Failure
func Fail(err error) Result {
	return Result{"", err.Error()}
}

// Success
func Done(data interface{}) Result {
	return Result{data, ""}
}

// Handle request
func handle(mdl, obj, act string, c reflect.Type) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf(r.(string))
				http.Error(w, r.(string), http.StatusBadRequest)
			}
		}()

		fmt.Print("\n\n")

		cp := reflect.New(c).Interface().(Controller)
		dump.Dump(cp)

		err := json.NewDecoder(r.Body).Decode(&cp)
		if err != nil {
			log.Printf("%v\n", err)
		}

		// Load()
		if err := cp.Load(mdl, obj, act, r); err != nil {
			panic(fmt.Sprintf("Load failed: %s\n", err))
		}

		// Execute action
		result := cp.Render()
		if result.Err != "" {
			panic(fmt.Sprintf("Execute error: %v\n", result.Err))
		}

		// JSON Marshal
		v, err := json.Marshal(result.Data)
		if err != nil {
			panic(fmt.Sprintf("json.Marshal: %v\n", err))
		}

		// Write to output
		_, err = w.Write(v)
		if err != nil {
			panic(fmt.Sprintf("ResponseWriter.Write: %v\n", err))
		}
	}
}

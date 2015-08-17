// Copyright 2014 The zhgo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package web

import (
	"encoding/json"
	"fmt"
	"github.com/zhgo/dump"
	"log"
	"net/http"
	"reflect"
	"strings"
)

// Controller interface
type Controller interface {
	Load(module, action string, r *http.Request) error
	Render() ActionResult
}

// Action result
type ActionResult struct {
	// Data
	Data interface{} `json:"data"`

	// error
	Err string `json:"err"`
}

func NewController(c Controller) {
	ct := reflect.TypeOf(c).Elem()

	la := strings.Split(ct.String(), ".")
	if len(la) != 2 {
		log.Fatalf("ControllerType.String() parase error: %s\n", ct.String())
	}

	module := strings.Title(la[0])
	action := la[1]
	pattern := fmt.Sprintf("/%s/%s", module, action)

	m, ok := app.Modules[module]
	if !ok {
		log.Fatalf("Module not found: %s\n", module)
	}

	if _, ok = app.muxList[m.Listen]; !ok {
		app.muxList[m.Listen] = http.NewServeMux()
	}

	app.muxList[m.Listen].HandleFunc(pattern, handle(module, action, ct))
}

// Failure
func Fail(err error) ActionResult {
	return ActionResult{"", err.Error()}
}

// Success
func Done(data interface{}) ActionResult {
	return ActionResult{data, ""}
}

func handle(module, action string, c reflect.Type) func(w http.ResponseWriter, r *http.Request) {
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
		if err := cp.Load(module, action, r); err != nil {
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

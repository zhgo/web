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
    Load(model, action string, r *http.Request) error
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
func NewController(c Controller, patterns ...string) {
    ct := reflect.TypeOf(c).Elem()

    strSli := strings.Split(ct.String(), ".")
    mdl := strSli[0][0 : len(strSli[0])-3]
    act := strSli[1] // ct.Name()

    m, ok := app.Modules[mdl]
    if !ok {
        // log.Fatalf("Module not found: %s\n", mdl)
        m = Module{mdl, app.Listen}
        app.Modules[mdl] = m
    }

    if _, ok = app.muxList[m.Listen]; !ok {
        app.muxList[m.Listen] = http.NewServeMux()
    }

    patterns = append(patterns, fmt.Sprintf("/%s/%s", mdl, act))
    for _, pattern := range patterns {
        log.Printf("%#v\n", pattern)
        app.muxList[m.Listen].HandleFunc(pattern, handle(mdl, act, ct))
    }
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
func handle(mdl, act string, c reflect.Type) func(w http.ResponseWriter, r *http.Request) {
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
        if err := cp.Load(mdl, act, r); err != nil {
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

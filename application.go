// Copyright 2014 The zhgo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package web

import (
	"encoding/json"
	"fmt"
	"github.com/zhgo/console"
	"github.com/zhgo/db"
	"log"
	"net/http"
	"reflect"
)

// Application struct
type Application struct {
	// Environment 0:development 1:testing 2:staging 3:production
	Environment int8

	// Listen address and port
	Listen string

	// Host list
	Hosts map[string]Host

	// module list
	Modules map[string]Module

	// *http.ServeMux
	muxList map[string]*http.ServeMux
}

// Module struct
type Module struct {
	// module name
	Name string

	// Listen
	Listen string

	// key of DB Server
	DB db.Server
}

// Host struct
type Host struct {
	//
	Name string

	// Listen
	Listen string

	// Path
	Path string

	// Root
	Root string
}

// Init
func (app *Application) Init(path string) {
	// Load config file
	r := map[string]string{"{WorkingDir}": console.WorkingDir}
	console.NewConfig(path).Replace(r).Parse(app)

	// Default Listen
	if app.Listen == "" {
		app.Listen = ":80"
	}

	// Default host
	if app.Hosts == nil {
		app.Hosts = make(map[string]Host)
	}
	if _, ok := app.Hosts["Public"]; !ok {
		app.Hosts["Public"] = Host{Path: "/", Root: console.WorkingDir + "/public"}
	}

	// Host property
	for k, v := range app.Hosts {
		v.Name = k
		if v.Listen == "" {
			v.Listen = app.Listen
		}
		app.Hosts[k] = v
	}

	// Default module
	if app.Modules == nil {
		app.Modules = make(map[string]Module)
	}

	// Module property
	for k, v := range app.Modules {
		// app.Modules[k].Listen = app.Listen // cannot assign to p.Modules[k].Listen
		v.Name = k
		if v.Listen == "" {
			v.Listen = app.Listen
		}

		// db.Connections
		if v.DB.Follow != "" {
			v.DB = app.Modules[v.DB.Follow].DB
		}

		if v.DB.DSN != "" {
			db.Servers[v.Name] = &v.DB
		}

		app.Modules[k] = v
	}

	app.muxList = make(map[string]*http.ServeMux)

	//log.Printf("%#v\n", app)
	console.Dump(app)
	//log.Printf("%#v\n", controllers)
	console.Dump(controllers)
}

// Load
func (app *Application) Load(fn ActionLoadFunc) {
	// hosts
	for _, m := range app.Hosts {
		if _, ok := app.muxList[m.Listen]; !ok {
			app.muxList[m.Listen] = http.NewServeMux()
		}

		if m.Path == "/" {
			app.muxList[m.Listen].Handle(m.Path, http.FileServer(http.Dir(m.Root)))
		} else {
			// To serve a directory on disk (/tmp) under an alternate URL
			// path (/tmpfiles/), use StripPrefix to modify the request
			// URL's path before the FileServer sees it:
			app.muxList[m.Listen].Handle(m.Path, http.StripPrefix(m.Path, http.FileServer(http.Dir(m.Root))))
		}
	}

	// modules
	for mName, m := range app.Modules {
		if _, ok := app.muxList[m.Listen]; !ok {
			app.muxList[m.Listen] = http.NewServeMux()
		}

		app.muxList[m.Listen].HandleFunc("/"+mName+"/", handleRequest(fn))
	}

	//log.Printf("%#v\n", app.muxList)
	console.Dump(app.muxList)
}

// Start HTPP server
func (app *Application) Start() {
	l := len(app.muxList)
	sem := make(chan int, l)

	for listen, mux := range app.muxList {
		go app.listenAndServe(listen, mux, sem)
	}

	for i := 0; i < l; i++ {
		<-sem
	}
}

//new host
func (app *Application) listenAndServe(listen string, mux *http.ServeMux, sem chan int) {
	err := http.ListenAndServe(listen, mux)
	if err != nil {
		log.Fatal(err)
	}
	sem <- 0
}

func valAction(req *Request) reflect.Value {
	cm, ok := controllers[req.Module]
	if !ok {
		panic(fmt.Sprintf("Controller [%s] not found\n", req.Module))
	}
	controller, ok := cm[req.Module+req.Controller]
	if !ok {
		panic(fmt.Sprintf("Controller [%s::%s] not found\n", req.Module, req.Controller))
	}

	rq := controller.Elem().FieldByName("Request")
	rq.Set(reflect.ValueOf(req))

	method := req.Action

	// Action
	action := controller.MethodByName(method)
	if action.IsValid() == false {
		panic(fmt.Sprintf("Method [%s] not found\n", method))
	}

	return action
}

func valArgs(req *Request, action reflect.Value) []reflect.Value {
	typ := action.Type()
	numIn := typ.NumIn()
	if len(req.args) != numIn {
		panic(fmt.Sprintf("Method [%s]'s in arguments wrong\n", req.Action))
	}

	// Arguments
	in := make([]reflect.Value, numIn)
	for i := 0; i < numIn; i++ {
		actionIn := typ.In(i)
		kind := actionIn.Kind()
		v, err := parameterConversion(req.args[i], kind)
		if err != nil {
			panic(fmt.Sprintf("%s paramters error, string convert to %s failed: %s\n", req.Action, kind, req.args[i]))
		}

		in[i] = v
		req.Args[actionIn.Name()] = v
	}

	return in
}

//Every request run this function
func handleRequest(fn ActionLoadFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf(r.(string))
				http.Error(w, r.(string), http.StatusBadRequest)
			}
		}()

		fmt.Print("\n\n")

		req := NewRequest(r)

		//log.Printf("%#v\n", req)
		console.Dump(req)

		// Invoke Load()
		if fn != nil {
			if code, err := fn(r, req); code < 0 {
				panic(fmt.Sprintf("Load failed: %s\n", err))
			}
		}

		action := valAction(req)

		log.Printf("Method [%s] found\n", req.Action)

		args := valArgs(req, action)

		// Execute action
		ret := action.Call(args)
		result := ret[0].Interface().(ActionResult)
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

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
	"runtime"
)

// Application struct
type Application struct {
	// Environment 0:development 1:testing 2:staging 3:production
	Environment int8

	// module list
	Modules map[string]Module

	// Listen address and port
	Listen string

	// Host list
	Hosts map[string]Host
}

// Module struct
type Module struct {
	// module name
	Name string

	// key of DB Server
	DB db.Server

	// Listen
	Listen string
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

	// Default module
	if app.Modules == nil {
		app.Modules = make(map[string]Module)
	}

	// Default host
	if app.Hosts == nil {
		app.Hosts = make(map[string]Host)
	}
	if _, ok := app.Hosts["Public"]; !ok {
		app.Hosts["Public"] = Host{Path: "/", Root: console.WorkingDir + "/public"}
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

	// Host property
	for k, v := range app.Hosts {
		v.Name = k
		if v.Listen == "" {
			v.Listen = app.Listen
		}
		app.Hosts[k] = v
	}

	//log.Printf("%#v\n", app)
	console.Dump(app)
	//log.Printf("%#v\n", controllers)
	console.Dump(controllers)
}

// Load
func (p *Application) Load(fn ActionLoadFunc) {
	// hosts
	for _, m := range p.Hosts {
		if _, ok := muxList[m.Listen]; !ok {
			muxList[m.Listen] = http.NewServeMux()
		}

		if m.Path == "/" {
			muxList[m.Listen].Handle(m.Path, http.FileServer(http.Dir(m.Root)))
		} else {
			// To serve a directory on disk (/tmp) under an alternate URL
			// path (/tmpfiles/), use StripPrefix to modify the request
			// URL's path before the FileServer sees it:
			muxList[m.Listen].Handle(m.Path, http.StripPrefix(m.Path, http.FileServer(http.Dir(m.Root))))
		}
	}

	// modules
	for mName, m := range p.Modules {
		if _, ok := muxList[m.Listen]; !ok {
			muxList[m.Listen] = http.NewServeMux()
		}

		muxList[m.Listen].HandleFunc("/"+mName+"/", NewHandler(fn))
	}
}

// Start HTPP server
func (p *Application) Start() {
	//log.Printf("%#v\n", muxList)

	runtime.GOMAXPROCS(runtime.NumCPU())
	l := len(muxList)
	sem := make(chan int, l)

	for listen, mux := range muxList {
		go p.listenAndServe(listen, mux, sem)
	}

	for i := 0; i < l; i++ {
		<-sem
	}
}

//new host
func (p *Application) listenAndServe(listen string, mux *http.ServeMux, sem chan int) {
	err := http.ListenAndServe(listen, mux)
	if err != nil {
		log.Fatal(err)
	}
	sem <- 0
}

//Every request run this function
func NewHandler(fn ActionLoadFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf(r.(string))
				http.Error(w, r.(string), http.StatusBadRequest)
			}
		}()

		fmt.Print("\n\n")

		req := NewRequest(r)
		console.Dump(req)

		cm, ok := controllers[req.Module]
		if !ok {
			panic(fmt.Sprintf("Controller [%s] not found\n", req.Module))
		}
		controller, ok := cm[req.Module+req.Controller]
		if !ok {
			panic(fmt.Sprintf("Controller [%s::%s] not found\n", req.Module, req.Controller))
		}

		//Invoke Load()
		if fn != nil {
			if inited := fn(req); inited.Data.(int) < 0 {
				panic(fmt.Sprintf("Load falure: %s\n", inited.Err))
			}
		}

		rq := controller.Elem().FieldByName("Request")
		rq.Set(reflect.ValueOf(req))

		method := req.Action

		//action
		action := controller.MethodByName(method)
		if action.IsValid() == false {
			panic(fmt.Sprintf("Method [%s] not found\n", method))
		}

		log.Printf("Method [%s] found\n", method)

		typ := action.Type()
		numIn := typ.NumIn()
		if len(req.args) != numIn {
			panic(fmt.Sprintf("Method [%s]'s in arguments wrong\n", method))
		}

		// Arguments
		in := make([]reflect.Value, numIn)
		for i := 0; i < numIn; i++ {
			actionIn := typ.In(i)
			kind := actionIn.Kind()
			v, err := parameterConversion(req.args[i], kind)
			if err != nil {
				panic(fmt.Sprintf("%s paramters error, string convert to %s failed: %s\n", method, kind, req.args[i]))
			}

			in[i] = v
			req.Args[actionIn.Name()] = v
		}

		// Run...
		ret := action.Call(in)

		result := ret[0].Interface().(Result)
		if result.Err != "" {
			panic(fmt.Sprintf("Execute error: %v\n", result.Err))
		}

		// JSON Marshal
		v, err := json.Marshal(result.Data)
		if err != nil {
			panic(fmt.Sprintf("json.Marshal: %v\n", err))
		}

		// Render
		_, err = w.Write(v)
		if err != nil {
			panic(fmt.Sprintf("ResponseWriter.Write: %v\n", err))
		}
	}
}

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

	// Mux map group by listen
	muxList map[string]*http.ServeMux
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

// Module struct
type Module struct {
	// module name
	Name string

	// Listen
	Listen string

	// key of DB Server
	DB db.Server
}

// Init
func (app *Application) Init() {
	// Load config file
	replaces := map[string]string{"{WorkingDir}": console.WorkingDir}
	console.NewConfig("zhgo.json").Replace(replaces).Parse(app)

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
}

// Load
func (app *Application) Load() {
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
}

// Start HTPP server
func (app *Application) Start() {
	//log.Printf("%#v\n", app)
	console.Dump(app)

	l := len(app.muxList)
	sem := make(chan int, l)

	for listen, mux := range app.muxList {
		go listenAndServe(listen, mux, sem)
	}

	for i := 0; i < l; i++ {
		<-sem
	}
}

//new host
func listenAndServe(listen string, mux *http.ServeMux, sem chan int) {
	err := http.ListenAndServe(listen, mux)
	if err != nil {
		log.Fatal(err)
	}
	sem <- 0
}

func Start() {
	App.Start()
}

// Failure
func Fail(err error) ActionResult {
	return ActionResult{"", err.Error()}
}

// Success
func Done(data interface{}) ActionResult {
	return ActionResult{data, ""}
}

func NewHandle(module string, pattern string, c Controller) {
	m, ok := App.Modules[module]
	if !ok {
		log.Fatalf("Module not found: %s\n", module)
	}

	if _, ok = App.muxList[m.Listen]; !ok {
		App.muxList[m.Listen] = http.NewServeMux()
	}

	App.muxList[m.Listen].HandleFunc(pattern, Handle(reflect.TypeOf(c).Elem()))
}

func Handle(c reflect.Type) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf(r.(string))
				http.Error(w, r.(string), http.StatusBadRequest)
			}
		}()

		fmt.Print("\n\n")

		cp := reflect.New(c).Interface().(Controller)

		console.Dump(cp)

		req := NewRequest(r)

		//log.Printf("%#v\n", req)
		console.Dump(req)

		err := json.NewDecoder(r.Body).Decode(&cp)
		if err != nil {
			log.Printf("%v\n", err)
		}

		// Load()
		if err := cp.Load(req); err != nil {
			panic(fmt.Sprintf("Load failed: %s\n", err))
		}

		// Execute action
		result := cp.Render(req)
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

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
	"strings"
)

// App
var App Application

// Application struct
type Application struct {
	// Environment 0:development 1:testing 2:staging 3:production
	Env int8 `json:"env"`

	// Listen address and port
	Listen string `json:"listen"`

	// Host list
	Hosts map[string]Host `json:"hosts"`

	// module list
	Modules map[string]Module `json:"modules"`

	// Mux map group by listen
	muxList map[string]*http.ServeMux
}

// Host struct
type Host struct {
	//
	Name string `json:"name"`

	// Listen
	Listen string `json:"listen"`

	// Path
	Path string `json:"path"`

	// Root
	Root string `json:"root"`
}

// Module struct
type Module struct {
	// module name
	Name string `json:"name"`

	// Listen
	Listen string `json:"listen"`

	// key of DB Server
	DB db.Server `json:"db"`
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

func NewHandle(c Controller) {
	ct := reflect.TypeOf(c).Elem()

	la := strings.Split(ct.String(), ".")
	if len(la) != 2 {
		log.Fatalf("ct.String() parase error: %s\n", ct.String())
	}

	module := strings.Title(la[0])
	action := la[1]
	pattern := fmt.Sprintf("/%s/%s", module, action)

	m, ok := App.Modules[module]
	if !ok {
		log.Fatalf("Module not found: %s\n", module)
	}

	if _, ok = App.muxList[m.Listen]; !ok {
		App.muxList[m.Listen] = http.NewServeMux()
	}

	App.muxList[m.Listen].HandleFunc(pattern, handle(module, action, ct))
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
		console.Dump(cp)

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

func Start() {
	App.Start()
}

func init() {
	App.Init()
	App.Load()
}

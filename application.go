// Copyright 2014 The zhgo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package web

import (
	"github.com/zhgo/config"
	"github.com/zhgo/kernel"
	"github.com/zhgo/db"
	"log"
	"net/http"
	"reflect"
	"strconv"
    "fmt"
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

//action result
type Result struct {
    //status
    Num int64

    //Message
    Msg string
}

// Action load function
type actionLoadFunc func(r *Request) Result

// App
var App Application

// *http.ServeMux
var muxList = make(map[string]*http.ServeMux)

// Init
func (app *Application) Init(path string) {
	// Load config file
	r := map[string]string{"{WorkingDir}": kernel.WorkingDir}
	config.LoadJSONFileAdv(app, path, r)

	// Default module
	if app.Modules == nil {
        app.Modules = make(map[string]Module)
	}

    // Default host
    if app.Hosts == nil {
        app.Hosts = make(map[string]Host)
        app.Hosts["Public"] = Host{Path: "/", Root: kernel.WorkingDir + "/public"}
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
            // v.DB.Follow = ""
        }
        if v.DB.DSN != "" {
            db.Servers[v.DB.Name] = &v.DB
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

	log.Printf("%#v\n", app)
	log.Printf("%#v\n", controllers)
}

// Load
func (p *Application) Load(fn actionLoadFunc) {
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

        muxList[m.Listen].HandleFunc("/"+mName+"/", newHandler(fn))
    }
}

// Start HTPP server
func (p *Application) Start() {
	//log.Printf("%#v\n", muxList)

    runtime.GOMAXPROCS(runtime.NumCPU())
	l := len(muxList)
    sem := make(chan int, l)

    for listen, mux := range muxList {
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

//Every request run this function
func newHandler(fn actionLoadFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if r := recover(); r != nil {
                log.Printf(r.(string))
            }
        }()

        req := NewRequest(r)
		log.Printf("\n\n%#v\n", req)

        cm, ok := controllers[req.Module];
		if !ok {
            panic(fmt.Sprintf("controller not found: %s\n", req.Module))
		}
		controller, ok := cm[req.Controller + "Controller"]
		if !ok {
            panic(fmt.Sprintf("controller not found: %s::%s\n", req.Module, req.Controller))
		}

		//Invoke Load()
		if fn != nil {
            if inited := fn(req); inited.Num < 0 {
                panic(fmt.Sprintf("Load falure: %s\n", inited.Msg))
            }
		}

		rq := controller.Elem().FieldByName("Request")
		rq.Set(reflect.ValueOf(req))

		view := NewView(req, w)
		vw := controller.Elem().FieldByName("View")
		vw.Set(reflect.ValueOf(view))

		method := req.Action

		//action
		action := controller.MethodByName(method)
		if action.IsValid() == false {
            panic(fmt.Sprintf("method [%s] not found\n", method))
		}

        log.Printf("method [%s] found\n", method)

        typ := action.Type()
        numIn := typ.NumIn()
        if len(req.args) < numIn {
            panic(fmt.Sprintf("method [%s]'s in arguments wrong\n", method))
        }

        in := make([]reflect.Value, numIn)
        for i := 0; i < numIn; i++ {
            actionIn := typ.In(i)
            kind := actionIn.Kind()
            v, err := parameterConversion(req.args[i], kind)
            if err != nil {
                panic(fmt.Sprintf("%s's paramters error, string convert to %s failed: %s\n", method, kind, req.args[i]))
            }

            in[i] = v
            req.Args[actionIn.Name()] = v
        }

        resultSli := action.Call(in)
        result := resultSli[0].Interface().(Result)
        view.realRender(result)
    }
}

// Parameter Conversion
func parameterConversion(str string, kind reflect.Kind) (reflect.Value, error) {
	var v reflect.Value
	var err error

	switch kind {
	case reflect.String:
		v = reflect.ValueOf(str)
	case reflect.Int64:
        var i64 int64
		i64, err = strconv.ParseInt(str, 10, 64)
		v = reflect.ValueOf(i64)
	default:
		log.Printf("string convert to int failure: %s\n", str)
	}

	return v, err
}

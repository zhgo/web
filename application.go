// Copyright 2014 The zhgo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package web

import (
	"github.com/zhgo/config"
	"github.com/zhgo/db"
	"log"
	"net/http"
	"reflect"
	"strconv"
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

    // database connection string
    //DB map[string]db.Config
}

// Module struct
type Module struct {
    // module name
    Name string

    // Listen
    Listen string

    // key of DSN
    DB db.DB
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

// Action load function
type actionLoadFunc func(r *Request) Result

// Root path
var WorkingDir string = config.WorkingDir()

// App
var App Application

func init() {
	App.Init(WorkingDir + "/example.json")
}

func (p *Application) Init(path string) {
	// load
	r := map[string]string{"{WorkingDir}": WorkingDir}
	config.LoadJSONFile(p, path, r)

	// default host
	if p.Hosts == nil {
		p.Hosts = make(map[string]Host)
		p.Hosts["Public"] = Host{Path: "/", Root: WorkingDir + "/public"}
	}

	// default module
	if p.Modules == nil {
		p.Modules = make(map[string]Module)
	}

	// default listen
	for k, v := range p.Modules {
		if v.Listen == "" {
			//v.Listen = p.Listen
			//p.Modules[k] = v

            // TODO: cannot assign to p.Modules[k].Listen
			p.Modules[k].Listen = p.Listen
		}
	}

	/*if p.DB == nil {
		p.DB = make(map[string]db.Config)
	}

    db.Configs = p.DB*/

	log.Printf("%#v\n", p)
}

// Start new HTPP server
func (p *Application) NewServer(fn actionLoadFunc) {
	muxList := make(map[string]*http.ServeMux)

	// hosts
	for _, m := range p.Hosts {
		_, s := muxList[m.Listen]
		if s == false {
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
		_, s := muxList[m.Listen]
		if s == false {
			muxList[m.Listen] = http.NewServeMux()
		}

        muxList[m.Listen].HandleFunc("/"+mName+"/", newHandler(fn))
	}

	//log.Printf("%#v\n", muxList)

	l := len(muxList)
	i := 0
	for listen, mux := range muxList {
		i++

		//log.Printf("%#v\n", mux)

		if i == l {
			//why the last listenning not go?
			newHost(listen, mux)
		} else {
			go newHost(listen, mux)
		}
	}
}

//new host
func newHost(listen string, mux *http.ServeMux) {
	err := http.ListenAndServe(listen, mux)
	if err != nil {
		log.Fatal(err)
	}
}

//Every request run this function
func newHandler(fn actionLoadFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := newRequest(r)

		log.Printf("\n\n") //空行分隔
		log.Printf("%#v\n", req)

		controller, s := controllers[req.Module+req.Controller]
		if s == false {
			log.Printf("controller not found: %s%s\n", req.Module, req.Controller)
			return
		}

		//Invoke Load() function
		if fn != nil {
			req.inited = fn(req)
			if req.inited.Num < 0 {
				req.Controller = "Index"
				req.Action = "Err"

				log.Printf("Load falure: %s\n", req.inited.Msg)
			}
		}

		rq := controller.Elem().FieldByName("Request")
		rq.Set(reflect.ValueOf(req))

		view := View{Request: req, Data: make(map[string]interface{}, 0), ResponseWriter: w}

		rw := controller.Elem().FieldByName("View")
		rw.Set(reflect.ValueOf(&view))

		method := req.Action

		//action
		action := controller.MethodByName(method)
		if action.IsValid() {
			log.Printf("method [%s] found\n", method)

			typ := action.Type()
			numIn := typ.NumIn()

			if len(req.args) >= numIn {
				pass := true
				in := make([]reflect.Value, numIn)

				for i := 0; i < numIn; i++ {
					actionIn := typ.In(i)
					kind := actionIn.Kind()
					v, err := paramConversion(kind, req.args[i])
					if err != nil {
						pass = false
						log.Printf("string convert to %s failure: %s\n", kind, req.args[i])
					} else {
						in[i] = v
						req.Args[actionIn.Name()] = v
					}
				}

				if pass == true {
					resultSli := action.Call(in)
					result := resultSli[0].Interface().(Result)
					//log.Printf("%#v\n", result)
					view.realRender(result)
				} else {
					log.Printf("%s's paramters failure\n", method)
				}
			} else {
				log.Printf("method [%s]'s in arguments wrong\n", method)
			}

		} else {
			log.Printf("method [%s] not found\n", method)
		}
	}
}

func paramConversion(kind reflect.Kind, arg string) (reflect.Value, error) {
	var v reflect.Value
	var err error

	switch kind {
	case reflect.String:
		v = reflect.ValueOf(arg)
	case reflect.Int64:
		var i64 int64
		i64, err = strconv.ParseInt(arg, 10, 64)
		v = reflect.ValueOf(i64)
	default:
		log.Printf("string convert to int failure: %s\n", arg)
	}

	return v, err
}

//action result
type Result struct {
	//status
	Num int64

	//Message
	Msg string
}

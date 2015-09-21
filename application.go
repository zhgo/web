// Copyright 2014 The zhgo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package web

import (
    "github.com/zhgo/config"
    "github.com/zhgo/dump"
    "log"
    "net/http"
)

// Application struct
type Application struct {
    // Environment 0:development 1:testing 2:staging 3:production
    Env int8 `json:"env"`

    // Host list
    Hosts map[string]Host `json:"hosts"`

    // Listen address and port
    Listen string `json:"listen"`

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
}

// Init
func (app *Application) Init() {
    // Load config file
    replaces := map[string]string{"{WorkingDir}": WorkingDir}
    config.NewConfig("zhgo.json").Replace(replaces).Parse(app)

    // Default Listen
    if app.Listen == "" {
        app.Listen = ":80"
    }

    // Default host
    if app.Hosts == nil {
        app.Hosts = make(map[string]Host)
    }

    // Host root
    if _, ok := app.Hosts["Public"]; !ok {
        app.Hosts["Public"] = Host{Path: "/", Root: WorkingDir + "/public"}
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
        if v.Name == "" {
            v.Name = k
        }

        // app.Modules[k].Listen = app.Listen // cannot assign to p.Modules[k].Listen
        if v.Listen == "" {
            v.Listen = app.Listen
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
    dump.Dump(app)

    l := len(app.muxList)
    sem := make(chan int, l)

    for listen, mux := range app.muxList {
        go listenAndServe(listen, mux, sem)
    }

    for i := 0; i < l; i++ {
        <-sem
    }
}

func Start() {
    app.Start()
}

//new host
func listenAndServe(listen string, mux *http.ServeMux, sem chan int) {
    err := http.ListenAndServe(listen, mux)
    if err != nil {
        log.Fatal(err)
    }
    sem <- 0
}

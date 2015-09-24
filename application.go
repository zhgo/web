// Copyright 2014 The zhgo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package web

import (
    "github.com/zhgo/config"
    "net/http"
)

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

// Init
func applicationInit() {
    // Load config file
    replaces := map[string]string{"{WorkingDir}": WorkingDir}
    config.NewConfig("zhgo.json").Replace(replaces).Parse(&app)

    // Default Listen
    if app.Listen == "" {
        app.Listen = ":80"
    }

    // Default host
    if app.Hosts == nil {
        app.Hosts = make(map[string]Host)
    }

    // Default module
    if app.Modules == nil {
        app.Modules = make(map[string]Module)
    }

    // Initalate muxList
    app.muxList = make(map[string]*http.ServeMux)

    // Host root
    if _, ok := app.Hosts["public"]; !ok {
        app.Hosts["public"] = Host{Path: "/", Root: WorkingDir + "/public"}
    }

    // Hosts
    for k, v := range app.Hosts {
        v.Name = k
        if v.Listen == "" {
            v.Listen = app.Listen
        }
        app.Hosts[k] = v
        NewHost(app.Hosts[k])
    }

    // Modules
    for k, v := range app.Modules {
        if v.Name == "" {
            v.Name = k
        }

        // cannot assign to p.Modules[k].Listen
        // app.Modules[k].Listen = app.Listen

        if v.Listen == "" {
            v.Listen = app.Listen
        }

        app.Modules[k] = v
    }
}

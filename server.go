// Copyright 2014 The zhgo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package web

import (
    "errors"
    "github.com/zhgo/dump"
    "log"
    "net/http"
)

// host lsitening
func listenAndServe(listen string, mux *http.ServeMux, sem chan int) {
    err := http.ListenAndServe(listen, mux)
    if err != nil {
        log.Fatal(err)
    }
    sem <- 0
}

func Start() {
    if started {
        log.Fatal(errors.New("Error, Server is running."))
    }

    //log.Printf("%#v\n", app)
    dump.Dump(app)

    l := len(app.muxList)
    sem := make(chan int, l)

    for listen, mux := range app.muxList {
        log.Printf("%#v\n", listen)
        go listenAndServe(listen, mux, sem)
    }

    started = true

    for i := 0; i < l; i++ {
        <-sem
    }

    started = false
}

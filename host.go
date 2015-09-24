// Copyright 2014 The zhgo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package web

import (
    "errors"
    "log"
    "net/http"
)

// Host struct
type Host struct {
    // Name
    Name string `json:"name"`

    // Listen
    Listen string `json:"listen"`

    // Path
    Path string `json:"path"`

    // Root
    Root string `json:"root"`
}

func NewHost(m Host) {
    if started {
        log.Fatal(errors.New("Cannot add host handle, Server is running."))
    }

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

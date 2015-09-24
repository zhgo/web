// Copyright 2014 The zhgo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package web

import (
    "github.com/zhgo/config"
)

// Module struct
type Module struct {
    // module name
    Name string `json:"name"`

    // Listen
    Listen string `json:"listen"`
}

// Action result
type Result struct {
    // Data
    Data interface{} `json:"data"`

    // error
    Err string `json:"err"`
}

// Working Directory
var WorkingDir string = config.WorkingDir()

// Initalation status
var started bool

// Application
var app Application

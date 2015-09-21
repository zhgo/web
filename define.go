// Copyright 2014 The zhgo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package web

import (
    "github.com/zhgo/config"
)

// Working Directory
var WorkingDir string = config.WorkingDir()

// Application
var app Application

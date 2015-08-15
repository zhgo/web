// Copyright 2014 The zhgo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package web

import (
	"net/http"
)

// Action load function
type ActionLoadFunc func(r *http.Request, req *Request) (int, string)

// App
var App Application

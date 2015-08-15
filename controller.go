// Copyright 2014 The zhgo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package web

import ()

// Controller interface
type Controller interface {
	Load(req *Request) error
	Render(req *Request) ActionResult
}

// Action result
type ActionResult struct {
	// Data
	Data interface{} `json:"data"`

	// error
	Err string `json:"err"`
}

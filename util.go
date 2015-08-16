// Copyright 2014 The zhgo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package web

import ()

// Failure
func Fail(err error) ActionResult {
	return ActionResult{"", err.Error()}
}

// Success
func Done(data interface{}) ActionResult {
	return ActionResult{data, ""}
}

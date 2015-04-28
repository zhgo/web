// Copyright 2014 The zhgo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package web

import (
	"testing"
)

func TestPathToMethod(t *testing.T) {
	path := "browse_by_set"
	method := pathToMethod(path)
	if method != "BrowseBySet" {
		t.Error("pathToMethod failure")
	}
}

func TestMethodToPath(t *testing.T) {
    method := "BrowseBySet"
    path := methodToPath(method)
    if path != "browse_by_set" {
        t.Errorf("methodToPath failure: %#v", path)
    }
}

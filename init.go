// Copyright 2014 The zhgo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package web

import (
	"log"
	"os"
)

func init() {
	var err error
	WorkingDir, err = os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	app.Init()

	app.Load()
}

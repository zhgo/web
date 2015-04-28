// Copyright 2014 The zhgo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package web

import (
	"fmt"
	"log"
)

func URL(m string, c string, a string, args ...interface{}) string {
	module, ok := App.Modules[m]
	if !ok {
		log.Printf("Module [%s] not exists\n", m)
		return ""
	}

    str := ""
    for _, v := range args {
        str += fmt.Sprintf("/%v", v)
    }

	return fmt.Sprintf("http://%s/%s/%s%s", module.Listen, c, a, str)
}

func Pub(p string) string {
	h, ok := App.Hosts["Public"]
	if !ok {
		log.Printf("Host [Public] not exists\n")
		return ""
	}

	return fmt.Sprintf("http://%s%s", h.Listen, p)
}

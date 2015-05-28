// Copyright 2014 The zhgo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package web

import (
    "html/template"
    "log"
    "reflect"
    "strconv"
    "strings"
    "unicode"
)

//action result
type Result struct {
    //Status code
    Code int

    //Message
    Err string
}

//funcMap
var funcMap = template.FuncMap{
    "FnURL": func(m string, c string, a string, args ...interface{}) string {
        return URL(m, c, a, args...)
    },
    "FnPub": func(p string) string {
        return Pub(p)
    },
}

// for example: transfer browse_by_set to BrowseBySet
func pathToMethod(path string) string {
    var method string
    sli := strings.Split(path, "_")
    for _, v := range sli {
        method += strings.Title(v)
    }
    return method
}

// for example: transfer BrowseBySet to browse_by_set
func methodToPath(method string) string {
    var words []string

    l := 0
    for s := method; s != ""; s = s[l:] {
        l = strings.IndexFunc(s[1:], unicode.IsUpper) + 1
        if l < 1 {
            l = len(s)
        }
        words = append(words, strings.ToLower(s[:l]))
    }

    return strings.Join(words, "_")
}

// Parameter Conversion
func parameterConversion(str string, kind reflect.Kind) (reflect.Value, error) {
    var v reflect.Value
    var err error

    switch kind {
        case reflect.String:
        v = reflect.ValueOf(str)
        case reflect.Int64:
        var i64 int64
        i64, err = strconv.ParseInt(str, 10, 64)
        v = reflect.ValueOf(i64)
        default:
        log.Printf("String convert to int failure: %s\n", str)
    }

    return v, err
}

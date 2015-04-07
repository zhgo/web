// Copyright 2014 The zhgo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"unicode"
    "github.com/zhgo/kernel"
)

//funcMaps
var funcMaps = template.FuncMap{
	"FavURL": func(m string, c string, a string, args ...interface{}) string {
		return Url(m, c, a, args...)
	},
	"FavMisc": func(a string) string {
		return a
	},
	"FavUI": func(a string) string {
		return a
	},
	"FavTheme": func(a string) string {
		return a
	},
}

//ResponseWriter struct
type View struct {
	//Request Refrence
	Request *Request

	//Page data
	Data map[string]interface{}

	//*http.ResponseWriter
	ResponseWriter http.ResponseWriter
}

//Set data
func (w *View) Set(key string, data interface{}) {
	w.Data[key] = data
}

//response html or json
func (w *View) Render() Result {
	return Result{1, ""}
}

//response html or json
func (w *View) realRender(ret Result) {
	switch w.Request.HTTPRequest.Method {
	case "GET": //output html
		log.Printf("%#v\n", w.Data)

		// new template
		t := template.New("Layout").Funcs(funcMaps)
        m := strings.ToLower(w.Request.Module)

		layoutPath := fmt.Sprintf("%s/web/%s/layout/normal.html", kernel.WorkingDir, m)
		t, err := t.ParseFiles(layoutPath)
		if err != nil {
			log.Printf("%s\n", err)
			return
		}

		// view html
		c := methodToPath(w.Request.Controller)
		a := methodToPath(w.Request.Action)
		viewPath := fmt.Sprintf("%s/web/%s/view/%s_%s.html", kernel.WorkingDir, m, c, a)

		log.Printf("%#v\n", viewPath)

		t, err = t.ParseFiles(viewPath)
		if err != nil {
			log.Printf("%s\n", err)
			return
		}

		//parse view
		err = t.Execute(w.ResponseWriter, w.Data)
		if err != nil {
			log.Printf("%s\n", err)
			return
		}
	default: //output json
		v, err := json.Marshal(ret)
		if err != nil {
			log.Println("json.Marshal:", err)
			return
		}

		_, err = w.ResponseWriter.Write(v)
		if err != nil {
			log.Println("w.ResponseWriter.Write:", err)
			return
		}
	}
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

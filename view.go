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
    "github.com/zhgo/kernel"
)

//View struct
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
func (w *View) render(ret Result) {
    log.Printf("%#v\n", w.Data)

    switch w.Request.HTTPRequest.Method {
        case "GET": //output html
        // new template
        t := template.New("Layout").Funcs(funcMap)
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

// New View
func NewView(r *Request, rw http.ResponseWriter) *View {
    return &View{Request: r, Data: make(map[string]interface{}, 0), ResponseWriter: rw}
}

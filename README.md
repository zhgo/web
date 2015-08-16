# Web --- Package for build web backend server (HTTP protocol, POST method, JSON format)

This is a special API server based on http protocol, only accept POST method, all incoming parameters are stored in the request body and must be JSON format, and can only output data in JSON format.

It's not like RESTful that allows GET, POST, PUT, DELETE and so on, it only allows POST methods.

# Feature

* Standalone server
* Only support POST method
* Only support JSON format
* Support multiple ports
* No custom routers
* One action is a struct, not a struct method

# Example

```go
package passport

import (
    "github.com/zhgo/web"
)

type UserList struct {
    
}

// This is only a example and keep it empty
func (c *UserList) Load(module, action string, r *http.Request) error {
    return nil
}

func (c *UserList) Render() web.ActionResult {
    d := map[string]string{"name": "John", "gender": "male"}
    list := []map[string]string{d}
    return web.Done(list)
}

func init() {
    web.NewHandle(new(UserList))
}
```

The UserList is action. Now you can write a main function call action above:

```go
package main

import (
    "github.com/zhgo/web"
    _ "passport" 
)

func main() {
    web.Start()
}
```

Then open browser and type this words in address input box:

http://localhost/Passport/UserList

```shell
[{"name": "John", "gender": "male"}]
```


[![Build Status](https://travis-ci.org/zhgo/web.svg)](https://travis-ci.org/zhgo/web)
[![Coverage Status](https://coveralls.io/repos/zhgo/web/badge.svg)](https://coveralls.io/r/zhgo/web)
[![GoDoc](https://godoc.org/github.com/zhgo/web?status.png)](http://godoc.org/github.com/zhgo/web)
[![License](https://img.shields.io/badge/license-BSD-blue.svg?style=flat)](https://github.com/zhgo/web/blob/master/LICENSE)

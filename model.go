// Copyright 2014 The zhgo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package web

import (
	"errors"
	"github.com/zhgo/db"
	"log"
)

// Module struct
type Module struct {
	// module name
	// Name string

	// module type
	// 1 is default, 2 is application
	Type int8

	// key of DSN
	DB string

	// Listen
	Listen string

	// Path
	Path string

	// Root
	Root string
}

// Model struct
type Model struct {
	// Module
	Module Module

	//table instance
	table db.Table
}

//Init
func (m *Model) Init(moduleName string, tableName string, t interface{}) {
	module, s := App.Modules[moduleName] //cannot get map element ptr directly
	if s == false {
		log.Fatal(errors.New("module config not found"))
	}

	primary, fields := db.TableFields(t)

	m.Module = module
	m.table = db.Table{Name: tableName, Primary: primary, Fields: fields}
}

//Same as fav.newQuery()
func (m *Model) Query() *db.Query {
	c, s := App.DB[m.Module.DB]
	if s == false {
		return nil //errors.New("DB config not found.")
	}
	return db.NewQuery(m.table, m.Module.DB, c)
}

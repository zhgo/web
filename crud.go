// Copyright 2014 The zhgo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package web

import (
	"encoding/json"
	"github.com/zhgo/console"
	"github.com/zhgo/db"
)

//controller struct
type Crud struct {
	//Request
	Request *Request
}

// Add
func (c *Crud) Add() interface{} {
	return c.crud(func(jsonPost db.Item, s *db.Server, tbName string, primary string) interface{} {
		r, err := s.InsertInto(tbName).Exec(jsonPost)
		if err != nil {
			return Result{-600, err.Error()}
		}

		return Result{r.LastInsertId, ""}
	})
}

// Detail
func (c *Crud) Detail(key string) interface{} {
	return c.crud(func(jsonPost db.Item, s *db.Server, tbName string, primary string) interface{} {
		d := db.Item{}
		q := s.NewQuery()
		err := q.Select("*").From(tbName).Where(q.Eq(primary, key)).Row(&d)
		if err != nil {
			return Result{-600, err.Error()}
		}

		return Result{d, ""}
	})
}

// Edit
func (c *Crud) Edit(key string) interface{} {
	return c.crud(func(jsonPost db.Item, s *db.Server, tbName string, primary string) interface{} {
		w := db.Where{primary: key}
		r, err := s.Update(tbName).Exec(jsonPost, w)
		if err != nil {
			return Result{-600, err.Error()}
		}

		return Result{r.RowsAffected, ""}
	})
}

// Delete
func (c *Crud) Delete(key string) interface{} {
	return c.crud(func(jsonPost db.Item, s *db.Server, tbName string, primary string) interface{} {
		w := db.Where{primary: key}
		r, err := s.DeleteFrom(tbName).Exec(w)
		if err != nil {
			return Result{-600, err.Error()}
		}

		return Result{r.RowsAffected, ""}
	})
}

// List
func (c *Crud) List(pi, ps int64) interface{} {
	return c.crud(func(jsonPost db.Item, s *db.Server, tbName string, primary string) interface{} {
		d := []db.Item{}
		q := s.NewQuery()
		err := q.Select("*").From(tbName).Rows(&d)
		if err != nil {
			return Result{-600, err.Error()}
		}

		return Result{d, ""}
	})
}

// crud
func (c *Crud) crud(fn func(jsonPost db.Item, s *db.Server, tbName string, primary string) interface{}) interface{} {
	jsonPost := db.Item{}
	err := json.NewDecoder(c.Request.HTTPRequest.Body).Decode(&jsonPost)
	if err != nil {
		return Result{-100, err.Error()}
	}

	if s, ok := db.Servers[App.Modules[c.Request.Module].DB.Name]; ok {
		tbName := console.CamelcaseToUnderscore(c.Request.Controller)
		primary := tbName + "_id"
		return fn(jsonPost, s, tbName, primary)
	} else {
		return Result{-101, "No database configuration"}
	}
}

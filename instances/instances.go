package instances

import (
	"encoding/json"
	"fmt"

	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/utility"
	"github.com/celsiainternet/jdb/jdb"
)

type Instance struct {
	schema *jdb.Schema
	model  *jdb.Model
}

var inst *Instance

/**
* Load
* @param db *jdb.DB, schema, name string
* @return (*Instance, error)
**/
func Load(db *jdb.DB, schema, name string) (*Instance, error) {
	if inst != nil {
		return inst, nil
	}

	var err error
	inst, err = Define(db, schema, name)
	if err != nil {
		return nil, err
	}

	return inst, nil
}

/**
* Define
* @param db *jdb.DB, schema, name string
* @return (*Instance, error)
**/
func Define(db *jdb.DB, schema, name string) (*Instance, error) {
	schemaObj, err := defineSchema(db, schema)
	if err != nil {
		return nil, err
	}

	if name == "" {
		name = "instances"
	}

	model := jdb.NewModel(schemaObj, name, 1)
	model.DefineCreatedAtField()
	model.DefineUpdatedAtField()
	model.DefineColumn(jdb.KEY, jdb.TypeDataKey)
	model.DefineColumn("tag", jdb.TypeDataKey)
	model.DefineColumn("definition", jdb.TypeDataBytes)
	model.DefinePrimaryKey(jdb.KEY)
	model.DefineSystemKeyField()
	model.DefineIndex(true, "tag", "kind", "user_id")

	if err := model.Init(); err != nil {
		return nil, err
	}

	return &Instance{
		schema: schemaObj,
		model:  model,
	}, nil
}

/**
* Get
* @param id string, dest any
* @return (bool, error)
**/
func (s *Instance) Get(id string, dest any) (bool, error) {
	if s.model == nil {
		return false, fmt.Errorf("model not found")
	}

	item, err := s.model.
		Where(jdb.KEY).Eq(id).
		One()
	if err != nil {
		return false, err
	}

	if !item.Ok {
		return false, nil
	}

	bt, err := item.Byte("definition")
	if err != nil {
		return false, err
	}

	err = json.Unmarshal(bt, &dest)
	if err != nil {
		return false, err
	}

	return true, nil
}

/**
* Set
* @param id string, tag string, obj any
* @return error
**/
func (s *Instance) Set(id, tag string, obj any) error {
	if s.model == nil {
		return nil
	}

	bt, ok := obj.([]byte)
	if !ok {
		var err error
		bt, err = json.Marshal(obj)
		if err != nil {
			return err
		}
	}

	data := et.Json{
		jdb.KEY:      id,
		"tag":        tag,
		"definition": bt,
	}

	now := utility.Now()
	_, err := s.model.
		Upsert(data).
		BeforeInsert(func(tx *jdb.Tx, data et.Json) error {
			data[jdb.CREATED_AT] = now
			data[jdb.UPDATED_AT] = now
			return nil
		}).
		BeforeUpdate(func(tx *jdb.Tx, data et.Json) error {
			data[jdb.UPDATED_AT] = now
			return nil
		}).
		Exec()
	if err != nil {
		return err
	}

	return nil
}

/**
* Delete
* @param id string
* @return error
**/
func (s *Instance) Delete(id string) error {
	if s.model == nil {
		return nil
	}

	_, err := s.model.
		Delete(jdb.KEY).Eq(id).
		Exec()
	if err != nil {
		return err
	}

	return nil
}

/**
* Query
* @param query et.Json
* @return et.Json, error
**/
func (s *Instance) Query(query et.Json) (et.Json, error) {
	result, err := jdb.From(s.model).
		Query(query)
	if err != nil {
		return et.Json{}, err
	}

	return result, nil
}

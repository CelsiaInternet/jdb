package instances

import (
	"encoding/json"
	"fmt"

	"github.com/celsiainternet/elvis/console"
	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/utility"
	"github.com/celsiainternet/jdb/jdb"
)

type Instance struct {
	schema *jdb.Schema
	model  *jdb.Model
}

var instance *Instance

func Define(db *jdb.DB, schema, name string) (*Instance, error) {
	if instance != nil {
		return instance, nil
	}

	instance = &Instance{}

	if err := instance.defineSchema(db, schema); err != nil {
		return nil, console.Panic(err)
	}

	if name == "" {
		name = "instances"
	}

	instance.model = jdb.NewModel(instance.schema, name, 1)
	instance.model.DefineCreatedAtField()
	instance.model.DefineUpdatedAtField()
	instance.model.DefineColumn(jdb.KEY, jdb.TypeDataKey)
	instance.model.DefineColumn("tag", jdb.TypeDataKey)
	instance.model.DefineColumn("definition", jdb.TypeDataBytes)
	instance.model.DefinePrimaryKey(jdb.KEY)
	instance.model.DefineSystemKeyField()
	instance.model.DefineIndexField()
	instance.model.DefineIndex(true, "tag")

	if err := instance.model.Init(); err != nil {
		return nil, err
	}

	return instance, nil
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

	items, err := s.model.
		Where(jdb.KEY).Eq(id).
		One()
	if err != nil {
		return false, err
	}

	if !items.Ok {
		return false, nil
	}

	scr, err := items.Byte("definition")
	if err != nil {
		return false, err
	}

	err = json.Unmarshal(scr, dest)
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

	now := utility.Now()
	data := et.Json{
		jdb.KEY:      id,
		"tag":        tag,
		"definition": bt,
	}

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

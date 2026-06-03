package authorization

import (
	"errors"
	"fmt"

	"github.com/celsiainternet/elvis/cache"
	"github.com/celsiainternet/elvis/dt"
	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/event"
	"github.com/celsiainternet/elvis/msg"
	"github.com/celsiainternet/elvis/timezone"
	"github.com/celsiainternet/elvis/utility"
	"github.com/celsiainternet/jdb/jdb"
)

type Authorization struct {
	schema *jdb.Schema
	model  *jdb.Model
}

var (
	auth        *Authorization
	ErrorInsert = fmt.Errorf(msg.RECORD_FOUND)
)

/**
* Load: Returns the existing Authorization singleton or creates one if it does not exist.
* @param db *jdb.DB, schema string, name string
* @return *Authorization, error
**/
func Load(db *jdb.DB, schema, name string) (*Authorization, error) {
	if auth != nil {
		return auth, nil
	}

	var err error
	auth, err = Define(db, schema, name)
	if err != nil {
		return nil, err
	}

	return auth, nil
}

/**
* Define: Creates and initializes a new Authorization instance with its schema and model.
* @param db *jdb.DB, schema string, name string
* @return *Authorization, error
**/
func Define(db *jdb.DB, schema, name string) (*Authorization, error) {
	_, err := cache.Load()
	if err != nil {
		return nil, err
	}

	schemaObj, err := defineSchema(db, schema)
	if err != nil {
		return nil, err
	}

	if name == "" {
		name = "Authorizationes"
	}

	model := jdb.NewModel(schemaObj, name, 1)
	model.DefineCreatedAtField()
	model.DefineSystemKeyField()
	model.DefineIndexField()
	model.DefineColumn("project_id", jdb.TypeDataKey)
	model.DefineColumn("profile_id", jdb.TypeDataKey)
	model.DefineColumn("method", jdb.TypeDataText)
	model.DefineColumn("path", jdb.TypeDataFullText)
	model.DefinePrimaryKey("project_id", "profile_id", "method", "path")
	model.BeforeInsert(func(tx *jdb.Tx, data et.Json) error {
		projectId := data.Str("project_id")
		profileId := data.Str("profile_id")
		method := data.Str("method")
		path := data.Str("path")
		exist, err := model.
			Where("project_id").Eq(projectId).
			And("profile_id").Eq(profileId).
			And("method").Eq(method).
			And("path").Eq(path).
			ItExistsTx(tx)
		if err != nil {
			return err
		}

		if exist {
			return ErrorInsert
		}

		return nil
	})

	if err := model.Init(); err != nil {
		return nil, err
	}

	return &Authorization{
		schema: schemaObj,
		model:  model,
	}, nil
}

/**
* Author: Checks whether a profile is authorized to access a given method and path within a project.
* @param projectId string, profileId string, method string, path string
* @return bool, error
**/
func (s *Authorization) Author(projectId, profileId, method, path string) (bool, error) {
	key := fmt.Sprintf("%s:%s:%s:%s", projectId, profileId, method, path)
	result := dt.Get(key)
	if result.Ok {
		return result.Bool("ok"), nil
	}

	ok, err := s.model.
		Where("project_id").Eq(projectId).
		And("profile_id").Eq(profileId).
		And("method").Eq(method).
		And("path").Eq(path).
		ItExists()
	if err != nil {
		return false, err
	}

	dt.Up(key, et.Item{Ok: ok, Result: et.Json{"ok": ok}})
	return ok, nil
}

/**
* RemoveAuthor: Removes the authorization for a profile to access a given method and path within a project.
* @param projectId string, profileId string, method string, path string
* @return error
**/
func (s *Authorization) RemoveAuthor(projectId, profileId, method, path string) error {
	key := fmt.Sprintf("%s:%s:%s:%s", projectId, profileId, method, path)
	dt.Drop(key)

	_, err := s.model.
		Delete("project_id").Eq(projectId).
		And("profile_id").Eq(profileId).
		And("method").Eq(method).
		And("path").Eq(path).
		Exec()
	if err != nil {
		return err
	}

	event.Publish(EVENT_DEL_AUTHORIZATION, et.Json{key: key})
	return nil
}

/**
* SetAuthor: Grants a profile authorization to access a given method and path within a project.
* @param projectId string, profileId string, method string, path string
* @return error
**/
func (s *Authorization) SetAuthor(projectId, profileId, method, path string) error {
	if !utility.ValidStr(method, 0, []string{""}) {
		return fmt.Errorf(msg.MSG_ATRIB_REQUIRED, "method")
	}
	if !utility.ValidStr(path, 0, []string{""}) {
		return fmt.Errorf(msg.MSG_ATRIB_REQUIRED, "path")
	}

	key := fmt.Sprintf("%s:%s:%s:%s", projectId, profileId, method, path)
	now := timezone.Now()
	_, err := s.model.
		Insert(et.Json{
			"project_id": projectId,
			"profile_id": profileId,
			"method":     method,
			"path":       path,
		}).
		BeforeInsert(func(tx *jdb.Tx, data et.Json) error {
			data.Set(jdb.CREATED_AT, now)
			return nil
		}).
		One()
	if err != nil {
		return err
	}

	dt.Drop(key)

	return nil
}

/**
* SetPath: Registers a method and path as a known authorization endpoint.
* @param method string, path string
* @return error
**/
func (s *Authorization) SetPath(method, path string) error {
	err := s.SetAuthor("", "", method, path)
	if err != nil && !errors.Is(err, ErrorInsert) {
		return err
	}

	return nil
}

/**
* RemovePath: Removes a method and path from the known authorization endpoints.
* @param method string, path string
* @return error
**/
func (s *Authorization) RemovePath(method, path string) error {
	_, err := s.model.
		Delete("method").Eq(method).
		And("path").Eq(path).
		Exec()
	if err != nil {
		return err
	}

	return nil
}

/**
* Query: Executes an authorization query and returns the result.
* @param query et.Json
* @return et.Json, error
**/
func (s *Authorization) Query(query et.Json) (et.Json, error) {
	result, err := jdb.From(s.model).
		Query(query)
	if err != nil {
		return nil, err
	}

	return result, nil
}

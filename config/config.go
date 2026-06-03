package config

import (
	"fmt"

	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/msg"
	"github.com/celsiainternet/jdb/jdb"
)

type Config struct {
	schema *jdb.Schema
	model  *jdb.Model
}

var (
	cfg         *Config
	ErrorInsert = fmt.Errorf(msg.RECORD_FOUND)
)

/**
* Load: Returns the existing Config singleton or creates one if it does not exist.
* @param db *jdb.DB, schema string, name string
* @return *Config, error
**/
func Load(db *jdb.DB, schema string) error {
	if cfg != nil {
		return nil
	}

	var err error
	cfg, err = Define(db, schema)
	if err != nil {
		return err
	}

	return nil
}

/**
* Define: Creates and initializes a new Config instance with its schema and model.
* @param db *jdb.DB, schema string, name string
* @return *Config, error
**/
func Define(db *jdb.DB, schema string) (*Config, error) {
	schemaObj, err := defineSchema(db, schema)
	if err != nil {
		return nil, err
	}

	name := "configs"
	model := jdb.NewModel(schemaObj, name, 1)
	model.DefineCreatedAtField()
	model.DefineSystemKeyField()
	model.DefineIndexField()
	model.DefineColumn("tag", jdb.TypeDataKey)
	model.DefineColumn("stage", jdb.TypeDataKey)
	model.DefineColumn("config", jdb.TypeDataObject)
	model.DefinePrimaryKey("tag", "stage")
	model.BeforeInsert(func(tx *jdb.Tx, data et.Json) error {
		tag := data.Str("tag")
		stage := data.Str("stage")
		exist, err := model.
			Where("tag").Eq(tag).
			And("stage").Eq(stage).
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

	return &Config{
		schema: schemaObj,
		model:  model,
	}, nil
}

/**
* Get: Returns the config object for a given tag and stage.
* @param tag string, stage string
* @return et.Json, error
**/
func (s *Config) Get(tag, stage string) (et.Item, error) {
	return s.model.
		Where("tag").Eq(tag).
		And("stage").Eq(stage).
		One()
}

/**
* Set: Inserts or updates the config object for a given tag and stage.
* @param tag string, stage string, config et.Json
* @return error
**/
func (s *Config) Set(tag, stage string, config et.Json) error {
	_, err := s.model.
		Upsert(et.Json{
			"tag":    tag,
			"stage":  stage,
			"config": config,
		}).
		Exec()
	return err
}

/**
* Delete: Removes the config record for a given tag and stage.
* @param tag string, stage string
* @return error
**/
func (s *Config) Delete(tag, stage string) error {
	_, err := s.model.
		Delete("tag").Eq(tag).
		And("stage").Eq(stage).
		Exec()
	return err
}

/**
* Query: Executes an authorization query and returns the result.
* @param query et.Json
* @return et.Json, error
**/
func (s *Config) Query(query et.Json) (et.Json, error) {
	result, err := jdb.From(s.model).
		Query(query)
	if err != nil {
		return nil, err
	}

	return result, nil
}

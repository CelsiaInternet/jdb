package config

import (
	"errors"

	"github.com/celsiainternet/elvis/et"
)

/**
* Get: Returns the config record for a given tag and stage.
* @param tag string, stage string
* @return et.Item, error
**/
func Get(tag, stage string) (et.Item, error) {
	if cfg == nil {
		return et.Item{}, errors.New(MSG_CONFIG_NOT_DEFINED)
	}

	return cfg.Get(tag, stage)
}

/**
* Set: Inserts or updates the config object for a given tag and stage.
* @param tag string, stage string, config et.Json
* @return error
**/
func Set(tag, stage string, config et.Json) error {
	if cfg == nil {
		return errors.New(MSG_CONFIG_NOT_DEFINED)
	}

	return cfg.Set(tag, stage, config)
}

/**
* Delete: Removes the config record for a given tag and stage.
* @param tag string, stage string
* @return error
**/
func Delete(tag, stage string) error {
	if cfg == nil {
		return errors.New(MSG_CONFIG_NOT_DEFINED)
	}

	return cfg.Delete(tag, stage)
}

/**
* Query: Executes a query against the config model and returns the result.
* @param query et.Json
* @return et.Json, error
**/
func Query(query et.Json) (et.Json, error) {
	if cfg == nil {
		return et.Json{}, errors.New(MSG_CONFIG_NOT_DEFINED)
	}

	return cfg.Query(query)
}

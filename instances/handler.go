package instances

import (
	"errors"

	"github.com/celsiainternet/elvis/et"
)

/**
* Get
* @param id string, dest any
* @return (bool, error)
**/
func Get(id string, dest any) (bool, error) {
	if inst == nil {
		return false, errors.New("instance not found")
	}

	return inst.Get(id, dest)
}

/**
* Set
* @param id string, tag string, obj any
* @return error
**/
func Set(id, tag string, obj any) error {
	if inst == nil {
		return errors.New("instance not found")
	}

	return inst.Set(id, tag, obj)
}

/**
* Delete
* @param id string
* @return error
**/
func Delete(id string) error {
	if inst == nil {
		return errors.New("instance not found")
	}

	return inst.Delete(id)
}

/**
* Query
* @param query et.Json
* @return et.Json, error
**/
func Query(query et.Json) (et.Json, error) {
	if inst == nil {
		return et.Json{}, errors.New("instance not found")
	}

	return inst.Query(query)
}

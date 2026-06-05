package authorization

import (
	"errors"

	"github.com/celsiainternet/elvis/et"
)

/**
* Author: Checks whether a profile is authorized to access a given method and path within a project.
* @param projectId string, profileId string, method string, path string
* @return bool, error
**/
func Author(projectId, profileId, method, path string) (bool, error) {
	if auth == nil {
		return false, errors.New(MSG_AUTHORIZATION_NOT_DEFINED)
	}

	return auth.Author(projectId, profileId, method, path)
}

/**
* RemoveAuthor: Removes the authorization of a profile to access a given method and path within a project.
* @param projectId string, profileId string, method string, path string
* @return error
**/
func RemoveAuthor(projectId, profileId, method, path string) error {
	if auth == nil {
		return errors.New(MSG_AUTHORIZATION_NOT_DEFINED)
	}

	return auth.RemoveAuthor(projectId, profileId, method, path)
}

/**
* SetAuthor: Grants a profile authorization to access a given method and path within a project.
* @param projectId string, profileId string, method string, path string
* @return error
**/
func SetAuthor(projectId, profileId, method, path string) error {
	if auth == nil {
		return errors.New(MSG_AUTHORIZATION_NOT_DEFINED)
	}

	return auth.SetAuthor(projectId, profileId, method, path)
}

/**
* SetPath: Registers a method and path as a known authorization endpoint.
* @param method string, path string
* @return error
**/
func SetPath(method, path string) error {
	if auth == nil {
		return errors.New(MSG_AUTHORIZATION_NOT_DEFINED)
	}

	return auth.SetPath(method, path)
}

/**
* RemovePath: Removes a method and path from the known authorization endpoints.
* @param method string, path string
* @return error
**/
func RemovePath(method, path string) error {
	if auth == nil {
		return errors.New(MSG_AUTHORIZATION_NOT_DEFINED)
	}

	return auth.RemovePath(method, path)
}

/**
* Query: Executes an authorization query and returns the result.
* @param query et.Json
* @return et.Json, error
**/
func Query(query et.Json) (et.Json, error) {
	if auth == nil {
		return et.Json{}, errors.New(MSG_AUTHORIZATION_NOT_DEFINED)
	}

	return auth.Query(query)
}

/**
* InitEvent: Initializes the authorization event listeners.
* @return error
**/
func InitEvent(projectId string, profiles []string) error {
	if auth == nil {
		return errors.New(MSG_AUTHORIZATION_NOT_DEFINED)
	}

	return auth.InitEvent(projectId, profiles)
}

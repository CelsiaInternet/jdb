package jdb

import (
	"fmt"

	"github.com/celsiainternet/elvis/et"
)

type Connected interface {
	Chain() (string, error)
	ToJson() et.Json
	Load(params et.Json) error
	Validate() error
}

type ConnectParams struct {
	Id       string    `json:"id"`
	Driver   string    `json:"driver"`
	HostName string    `json:"host_name"`
	Name     string    `json:"name"`
	IsDebug  bool      `json:"is_debug"`
	Params   Connected `json:"params"`
}

/**
* Json
* @return et.Json
**/
func (s *ConnectParams) ToJson() et.Json {
	return et.Json{
		"id":       s.Id,
		"driver":   s.Driver,
		"name":     s.Name,
		"is_debug": s.IsDebug,
		"params":   s.Params.ToJson(),
	}
}

/**
* LoadConnectParams
* @param params et.Json
* @return *ConnectParams, error
**/
func LoadConnectParams(params et.Json) (*ConnectParams, error) {
	connection := params.Json("params")
	result := &ConnectParams{
		Id:      params.Str("id"),
		Driver:  params.Str("driver"),
		Name:    params.Str("name"),
		IsDebug: params.Bool("is_debug"),
	}

	err := result.Params.Load(connection)
	if err != nil {
		return nil, err
	}

	return result, nil
}

/**
* Load
* @return *ConnectParams, error
**/
func load(driverName string) (*ConnectParams, error) {
	if driverName == "" {
		return nil, fmt.Errorf(MSG_DRIVER_NOT_DEFINED)
	}

	params, ok := conn.Params[driverName]
	if !ok {
		return nil, fmt.Errorf(MSG_DRIVER_NOT_DEFINED)
	}

	result := &ConnectParams{
		Id:       params.Id,
		Driver:   params.Driver,
		HostName: params.HostName,
		Name:     params.Name,
		IsDebug:  params.IsDebug,
		Params:   params.Params,
	}

	return result, nil
}

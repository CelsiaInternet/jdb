package jdb

import (
	"fmt"

	"github.com/celsiainternet/elvis/envar"
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
	Name     string    `json:"name"`
	UserCore bool      `json:"user_core"`
	NodeId   int       `json:"node_id"`
	Debug    bool      `json:"debug"`
	Params   Connected `json:"params"`
}

/**
* Json
* @return et.Json
**/
func (s *ConnectParams) ToJson() et.Json {
	return et.Json{
		"id":        s.Id,
		"driver":    s.Driver,
		"name":      s.Name,
		"user_core": s.UserCore,
		"node_id":   s.NodeId,
		"debug":     s.Debug,
		"params":    s.Params.ToJson(),
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
		Id:       params.Str("id"),
		Driver:   params.Str("driver"),
		Name:     params.Str("name"),
		UserCore: params.Bool("user_core"),
		NodeId:   params.Int("node_id"),
		Debug:    params.Bool("debug"),
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
func load() (*ConnectParams, error) {
	driverName := envar.GetStr(PostgresDriver, "DB_DRIVER")
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
		Name:     params.Name,
		UserCore: params.UserCore,
		NodeId:   params.NodeId,
		Debug:    params.Debug,
		Params:   params.Params,
	}

	return result, nil
}

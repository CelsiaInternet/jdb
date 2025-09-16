package jdb

import (
	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/event"
)

/**
* publishError
* @param model *Model, sql string, err error
**/
func publishError(model *Model, sql string, err error) {
	event.Publish(EVENT_MODEL_ERROR, et.Json{
		"db":     model.Db.Name,
		"schema": model.Schema,
		"model":  model.Name,
		"sql":    sql,
		"error":  err.Error(),
	})
}

/**
* publishInsert
* @param model *Model, sql string
**/
func publishCommand(command *Command) {
	model := command.getModel()
	sql := command.Sql
	commandName := command.Command.Str()
	event.Publish(EVENT_MODEL_INSERT, et.Json{
		"db":      model.Db.Name,
		"schema":  model.Schema,
		"model":   model.Name,
		"sql":     sql,
		"command": commandName,
	})
}

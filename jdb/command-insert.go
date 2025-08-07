package jdb

import (
	"errors"
	"fmt"

	"github.com/celsiainternet/elvis/et"
)

func (s *Command) inserted() error {
	if len(s.Data) == 0 {
		return fmt.Errorf(MSG_NOT_DATA, s.Command.Str(), s.From.Name)
	}

	if err := s.prepare(); err != nil {
		return err
	}

	model := s.From
	results, err := s.Db.Command(s)
	if err != nil {
		for _, event := range model.eventError {
			event(model, et.Json{
				"command": "insert",
				"sql":     s.Sql,
				"data":    s.Data,
				"error":   err.Error(),
			})
		}

		return err
	}

	s.Result = results
	if !s.Result.Ok {
		return errors.New(MSG_NOT_INSERT_DATA)
	}

	s.ResultMap, err = model.getMapResultByPk(s.Result.Result)
	if err != nil {
		return err
	}

	if !s.isSync && model.UseCore {
		model.Emit(EVENT_MODEL_SYNC, et.Json{
			"command": "insert",
			"db":      model.Db.Name,
			"schema":  model.Schema,
			"model":   model.Name,
			"sql":     s.Sql,
			"values":  s.Values,
		})
	}

	if !model.isAudit {
		audit("insert", s.Sql)
	}

	for _, after := range s.ResultMap {
		for _, event := range model.eventsInsert {
			err := event(s.tx, model, et.Json{}, after)
			if err != nil {
				return err
			}
		}

	}

	for _, data := range s.Data {
		for _, fn := range s.afterInsert {
			err := fn(s.tx, data)
			if err != nil {
				return err
			}
		}

	}

	return nil
}

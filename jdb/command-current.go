package jdb

import (
	"fmt"

	"github.com/celsiainternet/elvis/et"
)

func (s *Command) current(where et.Json) error {
	model := s.getModel()
	if model == nil {
		return fmt.Errorf(MSG_MODEL_REQUIRED)
	}

	if len(s.Data) != 1 {
		return fmt.Errorf(MSG_MANY_INSERT_DATA)
	}

	columns := model.getColumnsByType(TpColumn)
	ql := From(model)
	ql.setWheres(where)
	ql.setSelects(columns)
	current, err := ql.
		setDebug(s.IsDebug).
		AllTx(s.tx)
	if err != nil {
		return err
	}

	s.Current = current
	mapCurrent, err := model.getMapByPk(current.Result)
	if err != nil {
		return err
	}

	s.CurrentMap = mapCurrent

	return nil
}

package jdb

import (
	"fmt"

	"github.com/celsiainternet/elvis/et"
)

func (s *Command) getCurrent(data et.Json) error {
	model := s.getModel()
	if model == nil {
		return fmt.Errorf(MSG_MODEL_REQUIRED)
	}

	ql := From(model)
	if s.Command == Upsert {
		err := ql.getWhereByPrimaryKeys(data)
		if err != nil {
			return err
		}
	}
	for _, w := range s.Wheres {
		ql.AddWhere(w)
	}
	ql.IsDebug = s.IsDebug
	ql.language = s.language
	columns := model.getColumnsByType(TpColumn)
	ql.setSelects(columns...)
	current, err := ql.
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

package jdb

import (
	"fmt"
)

func (s *Command) current() error {
	model := s.getModel()
	if model == nil {
		return fmt.Errorf(MSG_MODEL_REQUIRED)
	}

	if len(s.Data) != 1 {
		return fmt.Errorf(MSG_MANY_INSERT_DATA)
	}

	columns := model.getColumnsByType(TpColumn)
	ql := From(model)
	ql.Wheres = append(ql.Wheres, s.Wheres...)
	ql.IsDebug = s.IsDebug
	ql.language = s.language
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

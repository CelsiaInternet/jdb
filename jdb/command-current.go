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

	ql := From(model)
	err := ql.getWhereByPrimaryKeys(s.Data[0])
	if err != nil {
		return err
	}
	for _, w := range s.Wheres {
		ql.Where(w.Field).Eq(w.Value)
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

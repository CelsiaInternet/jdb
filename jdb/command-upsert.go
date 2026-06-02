package jdb

import (
	"fmt"

	"github.com/celsiainternet/elvis/et"
)

func (s *Command) upsert() error {
	model := s.getModel()
	if model == nil {
		return fmt.Errorf(MSG_MODEL_REQUIRED)
	}

	if len(s.Data) != 1 {
		return fmt.Errorf(MSG_MANY_INSERT_DATA)
	}

	s.current()
	if s.Current.Ok {
		s.Command = Update
		s.getWhereByPrimaryKeys(s.Data[0])
		return s.updated()
	}

	s.Command = Insert
	return s.inserted()
}

/**
* getWhereByPrimaryKeys
* @param data et.Json
* @return error
**/
func (s *Command) getWhereByPrimaryKeys(data et.Json) error {
	for name, col := range s.From.Froms[0].PrimaryKeys {
		val := data.Get(name)
		if val == nil {
			return fmt.Errorf("getWhereByPrimaryKeys:"+MSG_PRIMARY_KEY_REQUIRED, name, s.From.Froms[0].Name, data.ToString())
		}

		s.Where(col.Name).Eq(val)
	}

	return nil
}

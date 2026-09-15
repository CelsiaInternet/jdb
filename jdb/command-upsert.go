package jdb

import (
	"fmt"

	"github.com/celsiainternet/elvis/et"
)

func (s *Command) upsert() error {
	if len(s.Data) != 1 {
		return fmt.Errorf(MSG_MANY_INSERT_DATA)
	}

	data := s.Data[0]
	current, err := s.getCurrent(data)
	if err != nil {
		return err
	}

	if !current.Ok {
		return s.inserted()
	}

	return s.updated(current)
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

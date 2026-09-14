package jdb

import (
	"fmt"

	"github.com/celsiainternet/elvis/et"
)

func (s *Command) inserted() error {
	if err := s.prepare(); err != nil {
		return err
	}

	model := s.getModel()
	if len(s.Data) == 0 {
		return fmt.Errorf(MSG_NOT_DATA, s.Command.Str(), model.Name)
	}

	results, err := s.Db.Command(s)
	if err != nil {
		return err
	}

	s.Result = results
	if !s.Result.Ok {
		return nil
	}

	for _, after := range results.Result {
		for _, fn := range model.afterInsert {
			err := fn(s.tx, after)
			if err != nil {
				return err
			}
		}
		for _, fn := range model.afterInsertTrigger {
			err := fn(s.tx, et.Json{}, after)
			if err != nil {
				return err
			}
		}
		for _, fn := range s.afterInsert {
			err := fn(s.tx, after)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

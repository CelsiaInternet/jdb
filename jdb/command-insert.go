package jdb

import (
	"github.com/celsiainternet/elvis/et"
)

func (s *Command) inserted() error {
	if err := s.prepare(); err != nil {
		return err
	}

	model := s.getModel()
	for _, data := range s.Data {
		s.New = data
		results, err := s.Db.Command(s)
		if err != nil {
			return err
		}

		s.Result = results
		if !s.Result.Ok {
			return nil
		}

		for _, new := range results.Result {
			for _, fn := range model.afterInsert {
				err := fn(s.tx, new)
				if err != nil {
					return err
				}
			}
			for _, fn := range model.afterInsertTrigger {
				err := fn(s.tx, et.Json{}, new)
				if err != nil {
					return err
				}
			}
			for _, fn := range s.afterInsert {
				err := fn(s.tx, new)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

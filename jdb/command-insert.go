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
		for _, fn := range model.beforeInsert {
			err := fn(s.tx, s.New)
			if err != nil {
				return err
			}
		}

		for _, fn := range model.beforeInsertTrigger {
			err := fn(s.tx, et.Json{}, s.New)
			if err != nil {
				return err
			}
		}

		for _, fn := range s.beforeInsert {
			err := fn(s.tx, s.New)
			if err != nil {
				return err
			}
		}

		for _, fn := range s.beforeInsertTrigger {
			err := fn(s.tx, et.Json{}, s.New)
			if err != nil {
				return err
			}
		}

		results, err := s.Db.Command(s)
		if err != nil {
			return err
		}

		s.Result = results
		if !s.Result.Ok {
			return nil
		}

		new := results.Result[0]
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

		for _, fn := range s.afterInsertTrigger {
			err := fn(s.tx, et.Json{}, new)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

package jdb

import "github.com/celsiainternet/elvis/et"

func (s *Command) deleted() error {
	if err := s.prepare(); err != nil {
		return err
	}

	model := s.getModel()
	results, err := s.Db.Command(s)
	if err != nil {
		return err
	}

	s.Result = results
	if !s.Result.Ok {
		return nil
	}

	for _, before := range results.Result {
		for _, fn := range model.afterDelete {
			err := fn(s.tx, before)
			if err != nil {
				return err
			}
		}
		for _, fn := range model.afterDeleteTrigger {
			err := fn(s.tx, before, et.Json{})
			if err != nil {
				return err
			}
		}
		for _, fn := range s.afterDelete {
			err := fn(s.tx, before)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

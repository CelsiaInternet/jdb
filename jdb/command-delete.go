package jdb

import "github.com/celsiainternet/elvis/et"

func (s *Command) deleted(current et.Items) error {
	if err := s.prepare(); err != nil {
		return err
	}

	model := s.getModel()

	for _, old := range current.Result {
		new := et.Json{}
		for _, fn := range model.beforeDelete {
			err := fn(s.tx, old)
			if err != nil {
				return err
			}
		}

		for _, fn := range model.beforeDeleteTrigger {
			err := fn(s.tx, old, new)
			if err != nil {
				return err
			}
		}

		for _, fn := range s.beforeDelete {
			err := fn(s.tx, old)
			if err != nil {
				return err
			}
		}

		for _, fn := range s.beforeDeleteTrigger {
			err := fn(s.tx, old, new)
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

		for _, fn := range model.afterDelete {
			err := fn(s.tx, new)
			if err != nil {
				return err
			}
		}

		for _, fn := range model.afterDeleteTrigger {
			err := fn(s.tx, old, new)
			if err != nil {
				return err
			}
		}

		for _, fn := range s.afterDelete {
			err := fn(s.tx, new)
			if err != nil {
				return err
			}
		}

		for _, fn := range s.afterDeleteTrigger {
			err := fn(s.tx, old, new)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

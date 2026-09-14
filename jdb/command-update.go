package jdb

import (
	"fmt"

	"github.com/celsiainternet/elvis/et"
)

func (s *Command) updated(current et.Items) error {
	if err := s.prepare(); err != nil {
		return err
	}

	model := s.getModel()
	if len(s.Data) != 1 {
		return fmt.Errorf(MSG_MANY_UPDATE_DATA)
	}

	for _, old := range current.Result {
		s.New = s.Data[0]
		for _, fn := range model.beforeUpdate {
			err := fn(s.tx, s.New)
			if err != nil {
				return err
			}
		}

		for _, fn := range model.beforeUpdateTrigger {
			err := fn(s.tx, old, s.New)
			if err != nil {
				return err
			}
		}

		for _, fn := range s.afterUpdate {
			err := fn(s.tx, s.New)
			if err != nil {
				return err
			}
		}

		for _, fn := range s.beforeUpdateTrigger {
			err := fn(s.tx, old, s.New)
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
		for _, fn := range model.afterUpdate {
			err := fn(s.tx, new)
			if err != nil {
				return err
			}
		}

		for _, fn := range model.afterUpdateTrigger {
			err := fn(s.tx, old, new)
			if err != nil {
				return err
			}
		}

		for _, fn := range s.afterUpdate {
			err := fn(s.tx, new)
			if err != nil {
				return err
			}
		}

		for _, fn := range s.afterUpdateTrigger {
			err := fn(s.tx, old, new)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

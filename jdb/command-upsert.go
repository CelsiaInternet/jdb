package jdb

import (
	"fmt"
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

package jdb

import "fmt"

func (s *Command) upsert() error {
	model := s.getModel()
	if model == nil {
		return fmt.Errorf(MSG_MODEL_REQUIRED)
	}

	if len(s.Data) != 1 {
		return fmt.Errorf(MSG_MANY_INSERT_DATA)
	}

	if len(s.QlWhere.Wheres) == 0 {
		data := s.Data[0]
		where, err := model.GetWhereByPrimaryKeys(data)
		if err != nil {
			return err
		}
		s.QlWhere.Wheres = append(s.QlWhere.Wheres, where...)
	}

	s.current()
	if s.Current.Ok {
		s.Command = Update
		return s.updated()
	}

	s.Command = Insert
	return s.inserted()
}

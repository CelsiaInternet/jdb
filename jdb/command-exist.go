package jdb

import (
	"errors"

	"github.com/celsiainternet/elvis/et"
)

func (s *Command) exists(model *Model, where et.Json) (bool, error) {
	if model == nil {
		return false, errors.New(MSG_MODEL_REQUIRED)
	}

	ql := From(model)
	ql.setWheres(where)
	exist, err := ql.
		setDebug(s.IsDebug).
		ItExistsTx(s.tx)
	if err != nil {
		return false, err
	}

	return exist, nil
}

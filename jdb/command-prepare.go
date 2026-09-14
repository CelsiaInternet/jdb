package jdb

import (
	"fmt"

	"github.com/celsiainternet/elvis/console"
)

/**
* prepare
* @return error
**/
func (s *Command) prepare() error {
	model := s.getModel()
	if model == nil {
		return fmt.Errorf(MSG_MODEL_REQUIRED)
	}

	if s.IsDebug {
		console.Debug(s.Describe().ToString())
	}

	return nil
}

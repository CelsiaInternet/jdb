package authorization

import (
	"github.com/celsiainternet/elvis/console"
	"github.com/celsiainternet/elvis/event"
	"github.com/celsiainternet/elvis/router"
)

const (
	EVENT_SET_AUTHORIZATION = "event:set:authorization"
	EVENT_DEL_AUTHORIZATION = "event:del:authorization"
)

/**
* InitEven: Subscribes to API gateway events for route authorization changes.
* @return void
**/
func (s *Authorization) InitEvent() error {
	err := event.Stack(router.APIGATEWAY_SET_RESOLVE, s.eventSetResolve)
	if err != nil {
		return err
	}

	err = event.Stack(router.APIGATEWAY_DELETE_RESOLVE, s.eventDeleteResolve)
	if err != nil {
		return err
	}

	return nil
}

/**
* eventSetResolve: Handles the set-resolve event to register a new authorized path.
* @param m event.EvenMessage
* @return void
**/
func (s *Authorization) eventSetResolve(m event.EvenMessage) {
	if m.MySelf {
		return
	}

	data := m.Data
	method := data.Str("method")
	path := data.Str("path")
	err := s.SetPath(method, path)
	if err != nil {
		console.AlertF(`Authorization gateway error:%s`, err.Error())
	}
}

/**
* eventDeleteResolve: Handles the delete-resolve event to remove an authorized path.
* @param m event.EvenMessage
* @return void
**/
func (s *Authorization) eventDeleteResolve(m event.EvenMessage) {
	if m.MySelf {
		return
	}

	data := m.Data
	method := data.Str("method")
	path := data.Str("path")
	err := s.RemovePath(method, path)
	if err != nil {
		console.AlertF(`Authorization gateway error:%s`, err.Error())
	}
}

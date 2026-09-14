package jdb

import (
	"fmt"

	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/reg"
	"github.com/celsiainternet/elvis/utility"
)

/**
* beforeInsertDefault
* @param tx *Tx, data et.Json
* @return error
**/
func (s *Command) beforeInsertDefault(tx *Tx, old, new et.Json) error {
	model := s.getModel()
	if model == nil {
		return fmt.Errorf(MSG_MODEL_REQUIRED)
	}

	if model.IndexField != nil && new.Int(model.IndexField.Name) == 0 {
		new[model.IndexField.Name] = reg.GenIndex()
	}

	if model.SystemKeyField != nil && new.Str(model.SystemKeyField.Name) == "" {
		new[model.SystemKeyField.Name] = model.GenId()
	}

	now := utility.Now()
	if model.CreatedAtField != nil && new.Str(model.CreatedAtField.Name) == "" {
		new[model.CreatedAtField.Name] = now
	}

	if model.UpdatedAtField != nil && new.Str(model.UpdatedAtField.Name) == "" {
		new[model.UpdatedAtField.Name] = now
	}

	return nil
}

/**
* beforeUpdateDefault
* @param tx *Tx, data et.Json
* @return error
**/
func (s *Command) beforeUpdateDefault(tx *Tx, old, new et.Json) error {
	model := s.getModel()
	if model == nil {
		return fmt.Errorf(MSG_MODEL_REQUIRED)
	}

	now := utility.Now()
	if model.CreatedAtField != nil {
		delete(new, model.CreatedAtField.Name)
	}

	if model.UpdatedAtField != nil && new.Str(model.UpdatedAtField.Name) == "" {
		new[model.UpdatedAtField.Name] = now
	}

	return nil
}

/**
* beforeDeleteDefault
* @param tx *Tx, data et.Json
* @return error
**/
func (s *Command) beforeDeleteDefault(tx *Tx, old, new et.Json) error {
	if s.From == nil {
		return fmt.Errorf(MSG_MODEL_REQUIRED)
	}

	return nil
}

/**
* BeforeInsert
* @param fn DataFunction
**/
func (s *Command) BeforeInsert(fn DataFunctionTx) *Command {
	s.beforeInsert = append(s.beforeInsert, fn)

	return s
}

/**
* BeforeUpdate
* @param fn DataFunction
**/
func (s *Command) BeforeUpdate(fn DataFunctionTx) *Command {
	s.beforeUpdate = append(s.beforeUpdate, fn)

	return s
}

/**
* BeforeDelete
* @param fn DataFunction
**/
func (s *Command) BeforeDelete(fn DataFunctionTx) *Command {
	s.beforeDelete = append(s.beforeDelete, fn)

	return s
}

/**
* BeforeInsertOrUpdate
* @param fn DataFunction
**/
func (s *Command) BeforeInsertOrUpdate(fn DataFunctionTx) *Command {
	s.beforeInsert = append(s.beforeInsert, fn)
	s.beforeUpdate = append(s.beforeUpdate, fn)

	return s
}

/**
* BeforeInsertOrUpdateTrigger
* @param fn TriggerFunction
**/
func (s *Command) BeforeInsertTrigger(fn TriggerFunctionTx) *Command {
	s.beforeInsertTrigger = append(s.beforeInsertTrigger, fn)

	return s
}

/**
* BeforeUpdateTrigger
* @param fn TriggerFunction
**/
func (s *Command) BeforeUpdateTrigger(fn TriggerFunctionTx) *Command {
	s.beforeUpdateTrigger = append(s.beforeUpdateTrigger, fn)

	return s
}

/**
* BeforeDeleteTrigger
* @param fn TriggerFunction
**/
func (s *Command) BeforeDeleteTrigger(fn TriggerFunctionTx) *Command {
	s.beforeDeleteTrigger = append(s.beforeDeleteTrigger, fn)

	return s
}

/**
* BeforeInsertOrUpdateTrigger
* @param fn TriggerFunction
**/
func (s *Command) BeforeInsertOrUpdateTrigger(fn TriggerFunctionTx) *Command {
	s.beforeInsertTrigger = append(s.beforeInsertTrigger, fn)
	s.beforeUpdateTrigger = append(s.beforeUpdateTrigger, fn)

	return s
}

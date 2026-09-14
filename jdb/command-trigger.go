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

/**
* AfterInsert
* @param fn DataFunction
* @return *Command
**/
func (s *Command) AfterInsert(fn DataFunctionTx) *Command {
	s.afterInsert = append(s.afterInsert, fn)

	return s
}

/**
* AfterUpdate
* @param fn DataFunctionTx
* @return *Command
**/
func (s *Command) AfterUpdate(fn DataFunctionTx) *Command {
	s.afterUpdate = append(s.afterUpdate, fn)

	return s
}

/**
* AfterDelete
* @param fn DataFunctionTx
* @return *Command
**/
func (s *Command) AfterDelete(fn DataFunctionTx) *Command {
	s.afterDelete = append(s.afterDelete, fn)

	return s
}

/**
* AfterInsertOrUpdate
* @param fn DataFunctionTx
* @return *Command
**/
func (s *Command) AfterInsertOrUpdate(fn DataFunctionTx) *Command {
	s.afterInsert = append(s.afterInsert, fn)
	s.afterUpdate = append(s.afterUpdate, fn)

	return s
}

/**
* AfterInsertTrigger
* @param fn TriggerFunction
**/
func (s *Command) AfterInsertTrigger(fn TriggerFunctionTx) *Command {
	s.afterInsertTrigger = append(s.afterInsertTrigger, fn)

	return s
}

/**
* AfterUpdateTrigger
* @param fn TriggerFunction
**/
func (s *Command) AfterUpdateTrigger(fn TriggerFunctionTx) *Command {
	s.afterUpdateTrigger = append(s.afterUpdateTrigger, fn)

	return s
}

/**
* AfterDeleteTrigger
* @param fn TriggerFunction
**/
func (s *Command) AfterDeleteTrigger(fn TriggerFunctionTx) *Command {
	s.afterDeleteTrigger = append(s.afterDeleteTrigger, fn)

	return s
}

/**
* AfterInsertOrUpdateTrigger
* @param fn TriggerFunction
**/
func (s *Command) AfterInsertOrUpdateTrigger(fn TriggerFunctionTx) *Command {
	s.afterInsertTrigger = append(s.afterInsertTrigger, fn)
	s.afterUpdateTrigger = append(s.afterUpdateTrigger, fn)

	return s
}

package jdb

import (
	"encoding/json"
	"fmt"

	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/utility"
)

type TypeCommand int

const (
	Insert TypeCommand = iota
	Update
	Delete
	Upsert
)

func (s TypeCommand) Str() string {
	switch s {
	case Insert:
		return "insert"
	case Update:
		return "update"
	case Delete:
		return "delete"
	case Upsert:
		return "upsert"
	default:
		return "No command"
	}
}

type Function func() error
type DataFunctionTx func(tx *Tx, new et.Json) error
type TriggerFunctionTx func(tx *Tx, old, new et.Json) error

type Command struct {
	*QlWhere
	Id                  string              `json:"id"`
	tx                  *Tx                 `json:"-"`
	Command             TypeCommand         `json:"command"`
	Db                  *DB                 `json:"-"`
	From                *QlFroms            `json:"-"`
	Data                []et.Json           `json:"data"`
	Current             []et.Json           `json:"-"`
	Result              et.Items            `json:"result"`
	Sql                 string              `json:"sql"`
	Args                []any               `json:"args"`
	beforeInsert        []DataFunctionTx    `json:"-"`
	beforeUpdate        []DataFunctionTx    `json:"-"`
	beforeDelete        []DataFunctionTx    `json:"-"`
	afterInsert         []DataFunctionTx    `json:"-"`
	afterUpdate         []DataFunctionTx    `json:"-"`
	afterDelete         []DataFunctionTx    `json:"-"`
	beforeInsertTrigger []TriggerFunctionTx `json:"-"`
	beforeUpdateTrigger []TriggerFunctionTx `json:"-"`
	beforeDeleteTrigger []TriggerFunctionTx `json:"-"`
	afterInsertTrigger  []TriggerFunctionTx `json:"-"`
	afterUpdateTrigger  []TriggerFunctionTx `json:"-"`
	afterDeleteTrigger  []TriggerFunctionTx `json:"-"`
}

/**
* NewCommand
* @param model *Model, data []et.Json, command TypeCommand
* @return *Command
**/
func NewCommand(model *Model, data []et.Json, command TypeCommand) *Command {
	result := &Command{
		Id:                  utility.UUID(),
		Command:             command,
		Db:                  model.Db,
		From:                newForms(),
		Data:                data,
		Current:             []et.Json{},
		beforeInsert:        []DataFunctionTx{},
		beforeUpdate:        []DataFunctionTx{},
		beforeDelete:        []DataFunctionTx{},
		afterInsert:         []DataFunctionTx{},
		afterUpdate:         []DataFunctionTx{},
		afterDelete:         []DataFunctionTx{},
		beforeInsertTrigger: []TriggerFunctionTx{},
		beforeUpdateTrigger: []TriggerFunctionTx{},
		beforeDeleteTrigger: []TriggerFunctionTx{},
		afterInsertTrigger:  []TriggerFunctionTx{},
		afterUpdateTrigger:  []TriggerFunctionTx{},
		afterDeleteTrigger:  []TriggerFunctionTx{},
		Args:                []any{},
		Result:              et.Items{},
	}
	result.From.add(model)
	result.QlWhere = newQlWhere()
	result.IsDebug = model.IsDebug
	result.beforeInsert = append(result.beforeInsert, result.beforeInsertDefault)
	result.beforeUpdate = append(result.beforeUpdate, result.beforeUpdateDefault)
	result.beforeDelete = append(result.beforeDelete, result.beforeDeleteDefault)

	for _, fn := range model.beforeInsert {
		result.beforeInsert = append(result.beforeInsert, fn)
	}

	for _, fn := range model.beforeUpdate {
		result.beforeUpdate = append(result.beforeUpdate, fn)
	}

	for _, fn := range model.beforeDelete {
		result.beforeDelete = append(result.beforeDelete, fn)
	}

	for _, fn := range model.afterInsert {
		result.afterInsert = append(result.afterInsert, fn)
	}

	for _, fn := range model.afterUpdate {
		result.afterUpdate = append(result.afterUpdate, fn)
	}

	for _, fn := range model.afterDelete {
		result.afterDelete = append(result.afterDelete, fn)
	}

	for _, fn := range model.beforeInsertTrigger {
		result.beforeInsertTrigger = append(result.beforeInsertTrigger, fn)
	}

	for _, fn := range model.beforeUpdateTrigger {
		result.beforeUpdateTrigger = append(result.beforeUpdateTrigger, fn)
	}

	for _, fn := range model.beforeDeleteTrigger {
		result.beforeDeleteTrigger = append(result.beforeDeleteTrigger, fn)
	}

	for _, fn := range model.afterInsertTrigger {
		result.afterInsertTrigger = append(result.afterInsertTrigger, fn)
	}

	for _, fn := range model.afterUpdateTrigger {
		result.afterUpdateTrigger = append(result.afterUpdateTrigger, fn)
	}

	for _, fn := range model.afterDeleteTrigger {
		result.afterDeleteTrigger = append(result.afterDeleteTrigger, fn)
	}

	return result
}

/**
* setTx
* @param tx *Tx
* @return *Command
**/
func (s *Command) setTx(tx *Tx) *Command {
	s.tx = tx

	return s
}

/**
* serialize
* @return []byte, error
**/
func (s *Command) serialize() ([]byte, error) {
	result, err := json.Marshal(s)
	if err != nil {
		return []byte{}, err
	}

	return result, nil
}

/**
* Describe
* @return et.Json
**/
func (s *Command) Describe() et.Json {
	definition, err := s.serialize()
	if err != nil {
		return et.Json{}
	}

	result := et.Json{}
	err = json.Unmarshal(definition, &result)
	if err != nil {
		return et.Json{}
	}

	result["command"] = s.Command.Str()
	result["wheres"] = s.getWheres()

	return result
}

/**
* Debug
* @param v bool
* @return *Command
**/
func (s *Command) Debug() *Command {
	s.QlWhere.Debug()
	return s
}

/**
* getModel
* @return *Model
**/
func (s *Command) getModel() *Model {
	return s.From.getModel(0)
}

/**
* getFrom
* @return *QlFrom
**/
func (s *Command) GetFrom() *QlFrom {
	return s.From.getForm(0)
}

/**
* getField
* @param name string
* @return *Field
**/
func (s *Command) getField(name string) *Field {
	return s.From.getField(name)
}

/**
* getCurrent
* @param data et.Json
* @return et.Items, error
**/
func (s *Command) getCurrent(data et.Json) (et.Items, error) {
	model := s.getModel()
	if model == nil {
		return et.Items{}, fmt.Errorf(MSG_MODEL_REQUIRED)
	}

	ql := From(model)
	if s.Command == Upsert {
		err := ql.getWhereByPrimaryKeys(data)
		if err != nil {
			return et.Items{}, err
		}
	}
	for _, w := range s.Wheres {
		ql.addCondition(w)
	}
	ql.IsDebug = s.IsDebug
	current, err := ql.
		AllTx(s.tx)
	if err != nil {
		return et.Items{}, err
	}

	return current, nil
}

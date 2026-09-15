package jdb

import (
	"encoding/json"
	"fmt"

	"github.com/celsiainternet/elvis/console"
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
	New                 et.Json             `json:"new"`
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
		New:                 et.Json{},
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
	result.beforeInsertTrigger = append(result.beforeInsertTrigger, result.beforeInsertDefault)
	result.beforeUpdateTrigger = append(result.beforeUpdateTrigger, result.beforeUpdateDefault)
	result.beforeDeleteTrigger = append(result.beforeDeleteTrigger, result.beforeDeleteDefault)
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
* setDebug
* @param debug bool
* @return *Command
**/
func (s *Command) setDebug(debug bool) *Command {
	s.IsDebug = debug
	return s
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
	return s.From.getFromByIndex(0)
}

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

/**
* Tx
* @return *Tx
**/
func (s *Command) Tx() *Tx {
	return s.tx
}

/**
* ExecTx
* @param tx *Tx
* @return et.Items, error
**/
func (s *Command) ExecTx(tx *Tx) (et.Items, error) {
	var err error
	if tx == nil {
		tx = NewTx()

		defer func() (et.Items, error) {
			if err == nil {
				err = tx.Commit()
				if err != nil {
					return et.Items{}, err
				}
			}

			return s.Result, err
		}()
	}

	s.setTx(tx)
	switch s.Command {
	case Insert:
		err := s.inserted()
		if err != nil {
			return et.Items{}, err
		}
	case Update:
		current, err := s.getCurrent(et.Json{})
		if err != nil {
			return et.Items{}, err
		}
		err = s.updated(current)
		if err != nil {
			return et.Items{}, err
		}
	case Delete:
		current, err := s.getCurrent(et.Json{})
		if err != nil {
			return et.Items{}, err
		}
		err = s.deleted(current)
		if err != nil {
			return et.Items{}, err
		}
	case Upsert:
		err := s.upsert()
		if err != nil {
			return et.Items{}, err
		}
	default:
		return et.Items{}, fmt.Errorf(MSG_NOT_COMMAND)
	}

	return s.Result, nil
}

/**
* OneTx
* @param tx *Tx
* @return et.Item, error
**/
func (s *Command) OneTx(tx *Tx) (et.Item, error) {
	result, err := s.ExecTx(tx)
	if err != nil {
		return et.Item{}, err
	}

	return result.First(), nil
}

/**
* Exec
* @return et.Items, error
**/
func (s *Command) Exec() (et.Items, error) {
	return s.ExecTx(nil)
}

/**
* One
* @return et.Item, error
**/
func (s *Command) One() (et.Item, error) {
	return s.OneTx(nil)
}

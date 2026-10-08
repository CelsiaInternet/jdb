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
type DataFunction func(new et.Json)
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
	Returning           []*Field            `json:"returning"`
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
		Result:              et.Items{},
		Returning:           []*Field{},
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
	}
	result.From.add(model)
	result.QlWhere = newQlWhere()
	result.IsDebug = model.IsDebug
	result.beforeInsertTrigger = append(result.beforeInsertTrigger, result.beforeInsertDefault)
	result.beforeUpdateTrigger = append(result.beforeUpdateTrigger, result.beforeUpdateDefault)
	result.beforeDeleteTrigger = append(result.beforeDeleteTrigger, result.beforeDeleteDefault)
	return result
}

func (s *Command) beginTx() (*Tx, error) {
	if s.tx != nil {
		return s.tx, nil
	}

	var err error
	s.tx, err = newTx(s.Db.Db)
	return s.tx, err
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
* getField
* @param name string
* @return *Field
**/
func (s *Command) getField(name string) *Field {
	return s.From.getField(name)
}

/**
* getReturns
* @return []*Field
**/
func (s *Command) Returns(fields ...interface{}) *Command {
	model := s.getModel()
	if model == nil {
		return s
	}

	for _, name := range fields {
		switch v := name.(type) {
		case string:
			field := s.getField(v)
			if field != nil {
				s.Returning = append(s.Returning, field)
			}
		case *Column:
			field := s.getField(v.Name)
			if field != nil {
				s.Returning = append(s.Returning, field)
			}
		case Column:
			field := s.getField(v.Name)
			if field != nil {
				s.Returning = append(s.Returning, field)
			}
		case *Field:
			s.Returning = append(s.Returning, v)
		case Field:
			s.Returning = append(s.Returning, &v)
		}
	}

	return s
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
* getCurrent
* @param data et.Json
* @return et.Items, error
**/
func (s *Command) getCurrent(data et.Json) (et.Items, *QlWhere, error) {
	model := s.getModel()
	if model == nil {
		return et.Items{}, nil, fmt.Errorf(MSG_MODEL_REQUIRED)
	}

	ql := From(model)
	ql.Froms.Froms[0].As = ""
	// UPDATE y DELETE no declaran alias de tabla: cada campo de s.Wheres guarda su propia copia del
	// QlFrom con el alias del comando, así que se limpia para que el SQL no referencie un alias inexistente
	for _, w := range s.Wheres {
		for _, value := range []interface{}{w.Field, w.Value} {
			if field, ok := value.(*Field); ok && field.Model != nil {
				field.Model.As = ""
			}
		}
	}
	// Update y Delete llaman con data vacío y filtran solo por s.Wheres; Upsert filtra por la PK de data
	if len(data) > 0 {
		err := ql.getWhereByPrimaryKeys(data)
		if err != nil {
			return et.Items{}, nil, err
		}
	} else if len(s.Wheres) == 0 {
		return et.Items{}, nil, fmt.Errorf("where is required in model:%s", model.Name)
	}
	for _, w := range s.Wheres {
		ql.AddCondition(w)
	}
	ql.IsDebug = s.IsDebug
	current, err := ql.
		AllTx(s.tx)
	if err != nil {
		return et.Items{}, nil, err
	}

	return current, ql.QlWhere, nil
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
		tx, err = s.beginTx()
		if err != nil {
			return et.Items{}, err
		}

		defer func() (et.Items, error) {
			if err != nil {
				tx.rollback()
				return et.Items{}, err
			}

			err = tx.commit()
			if err != nil {
				return et.Items{}, err
			}

			return s.Result, err
		}()
	} else if s.tx == nil && tx != nil {
		s.tx = tx
	}

	switch s.Command {
	case Insert:
		err = s.inserted()
		if err != nil {
			return et.Items{}, err
		}
	case Update:
		current, qlWhere, err := s.getCurrent(et.Json{})
		if err != nil {
			return et.Items{}, err
		}
		s.QlWhere = qlWhere
		err = s.updated(current)
		if err != nil {
			return et.Items{}, err
		}
	case Delete:
		current, qlWhere, err := s.getCurrent(et.Json{})
		if err != nil {
			return et.Items{}, err
		}
		s.QlWhere = qlWhere
		err = s.deleted(current)
		if err != nil {
			return et.Items{}, err
		}
	case Upsert:
		err = s.upsert()
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

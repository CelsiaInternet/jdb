package jdb

import (
	"fmt"
	"sync"

	"github.com/celsiainternet/elvis/envar"
	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/utility"
)

type QlFrom struct {
	*Model
	As string
}

type QlFroms struct {
	Froms []*QlFrom
	index int
}

func (s *QlFroms) ToJson() et.Json {
	result := et.Json{}
	for _, from := range s.Froms {
		result[from.As] = from.Model.Name
	}
	return result
}

/**
* getField
* @param name string
* @return *Field
**/
func (s *QlFroms) getField(name string) *Field {
	for _, from := range s.Froms {
		result := from.getField(name, false)
		if result != nil {
			result.As = from.As
			return result
		}
	}

	return nil
}

/**
* newForms
* @return *QlFroms
**/
func newForms() *QlFroms {
	return &QlFroms{
		Froms: make([]*QlFrom, 0),
		index: 65,
	}
}

/**
* add
* @param m *Model
* @return *QlFrom
**/
func (s *QlFroms) add(m *Model) *QlFrom {
	as := string(rune(s.index))
	from := &QlFrom{
		Model: m,
		As:    as,
	}

	s.Froms = append(s.Froms, from)
	s.index++

	return from
}

/**
* getModel
* @param idx int
* @return *Model
**/
func (s *QlFroms) getModel(idx int) *Model {
	if s.Froms[idx] == nil {
		return nil
	}

	return s.Froms[idx].Model
}

/**
* getForm
* @param idx int
* @return *QlFrom
**/
func (s *QlFroms) getForm(idx int) *QlFrom {
	return s.Froms[idx]
}

type TypeSelect int

const (
	Select TypeSelect = iota
	Source
)

type Ql struct {
	*QlWhere
	Id           string                    `json:"id"`
	Db           *DB                       `json:"-"`
	TypeSelect   TypeSelect                `json:"type_select"`
	Froms        *QlFroms                  `json:"froms"`
	Selects      []interface{}             `json:"selects"`
	Rollups      []*Rollup                 `json:"rollups"`
	Joins        []*QlJoin                 `json:"joins"`
	Hiddens      []string                  `json:"hiddens"`
	Details      map[string]*Relation      `json:"details"`
	CalcFunction map[string]DataFunctionTx `json:"-"`
	Groups       []*Field                  `json:"group_bys"`
	Havings      *QlWhere                  `json:"havings"`
	Orders       *QlOrder                  `json:"orders"`
	Sheet        int                       `json:"sheet"`
	Offset       int                       `json:"offset"`
	Limit        int                       `json:"limit"`
	Sql          string                    `json:"sql"`
	tx           *Tx                       `json:"-"`
	wg           *sync.WaitGroup           `json:"-"`
}

/**
* From
* @param model *Model
* @return *Ql
**/
func From(name interface{}) *Ql {
	var model *Model
	switch v := name.(type) {
	case *Model:
		model = v
	default:
		str := fmt.Sprintf("%v", v)
		model = GetModel(str)
	}

	tpSelect := Select
	if model.SourceField != nil {
		tpSelect = Source
	}

	result := &Ql{
		Id:           utility.UUID(),
		Db:           model.Db,
		TypeSelect:   tpSelect,
		Froms:        newForms(),
		Selects:      make([]interface{}, 0),
		Rollups:      make([]*Rollup, 0),
		Joins:        make([]*QlJoin, 0),
		Hiddens:      make([]string, 0),
		Details:      make(map[string]*Relation, 0),
		CalcFunction: make(map[string]DataFunctionTx, 0),
		Groups:       make([]*Field, 0),
		Offset:       0,
		Sheet:        0,
		wg:           &sync.WaitGroup{},
	}
	result.QlWhere = newQlWhere()
	result.IsDebug = model.IsDebug
	result.Havings = newQlWhere()
	result.Froms.add(model)
	max := envar.GetInt(1000, "DB_RECORD_LIMIT")
	if result.Limit > max {
		result.Limit = max
	}

	return result
}

/**
* Describe
* @return et.Json
**/
func (s *Ql) Describe() et.Json {
	return et.Json{
		"from":     s.Froms.ToJson(),
		"join":     s.getJoins(),
		"where":    s.getWheres(),
		"group_by": s.getGroupsBy(),
		"having":   s.Havings.getWheres(),
		"order_by": s.getOrders(),
		"select":   s.getSelects(),
		"limit":    s.getLimit(),
		"sql":      s.Sql,
	}
}

/**
* setTx
* @param tx *Tx
* @return *Ql
**/
func (s *Ql) setTx(tx *Tx) *Ql {
	s.tx = tx

	return s
}

/**
* Tx
* @return *Tx
**/
func (s *Ql) Tx() *Tx {
	return s.tx
}

/**
* setDebug
* @param value bool
* @return *Ql
**/
func (s *Ql) setDebug(value bool) *Ql {
	s.IsDebug = value
	return s
}

/**
* Debug
* @param v bool
* @return *Ql
**/
func (s *Ql) Debug() *Ql {
	s.QlWhere.Debug()
	return s
}

package jdb

import (
	"fmt"
	"sync"

	"github.com/celsiainternet/elvis/et"
)

type QlFrom struct {
	*Model
	As string
}

type QlFroms struct {
	Froms []*QlFrom
	index int
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
	Id         string          `json:"id"`
	Db         *DB             `json:"-"`
	TypeSelect TypeSelect      `json:"type_select"`
	Froms      *QlFroms        `json:"froms"`
	Joins      []*QlJoin       `json:"joins"`
	Selects    []*Field        `json:"selects"`
	Hiddens    []string        `json:"hiddens"`
	Details    []*Field        `json:"details"`
	Groups     []*Field        `json:"group_bys"`
	Havings    *QlHaving       `json:"havings"`
	Orders     *QlOrder        `json:"orders"`
	Concurrent []*Field        `json:"concurrent"`
	Sheet      int             `json:"sheet"`
	Offset     int             `json:"offset"`
	Limit      int             `json:"limit"`
	Sql        string          `json:"sql"`
	Help       et.Json         `json:"help"`
	tx         *Tx             `json:"-"`
	wg         *sync.WaitGroup `json:"-"`
}

/**
* Describe
* @return et.Json
**/
func (s *Ql) Describe() et.Json {
	return et.Json{
		"from":     s.getForms(),
		"join":     s.getJoins(),
		"where":    s.getWheres(),
		"group_by": s.getGroupsBy(),
		"having":   s.getHavings(),
		"order_by": s.getOrders(),
		"select":   s.getSelects(),
		"limit":    s.getLimit(),
		"sql":      s.Sql,
		"help":     s.Help,
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
* getField
* @param name string
* @return *Field
**/
func (s *Ql) getField(name string) *Field {
	for _, from := range s.Froms.Froms {
		result := from.getField(name, false)
		if result != nil {
			return result
		}
	}

	return nil
}

/**
* Where
* @param fld interface{}
* @return *Ql
**/
func (s *Ql) Where(fld interface{}) *Ql {
	if s.QlWhere == nil {
		s.QlWhere = newQlWhere()
	}
	s.QlWhere.Where(fld)
	return s
}

/**
* And
* @param val interface{}
* @return *Ql
**/
func (s *Ql) And(fld interface{}) *Ql {
	s.Where(fld)
	return s
}

/**
* Or
* @param fld interface{}
* @return *Ql
**/
func (s *Ql) Or(fld interface{}) *Ql {
	if s.QlWhere == nil {
		s.QlWhere = newQlWhere()
	}
	s.QlWhere.Or(fld)
	return s
}

/**
* Eq
* @param val interface{}
* @return *Ql
**/
func (s *Ql) Eq(val interface{}) *Ql {
	s.QlWhere.Eq(val)
	return s
}

/**
* Neg
* @param val interface{}
* @return *Ql
**/
func (s *Ql) Neg(val interface{}) *Ql {
	s.QlWhere.Neg(val)
	return s
}

/**
* In
* @param val ...any
* @return *Ql
**/
func (s *Ql) In(val ...any) *Ql {
	s.QlWhere.In(val...)
	return s
}

/**
* NotIn
* @param val ...any
* @return *Ql
**/
func (s *Ql) NotIn(val ...any) *Ql {
	s.QlWhere.NotIn(val...)
	return s
}

/**
* Like
* @param val interface{}
* @return *Ql
**/
func (s *Ql) Like(val interface{}) *Ql {
	s.QlWhere.Like(val)
	return s
}

/**
* More
* @param val interface{}
* @return *Ql
**/
func (s *Ql) More(val interface{}) *Ql {
	s.QlWhere.More(val)
	return s
}

/**
* Less
* @param val interface{}
* @return *Ql
**/
func (s *Ql) Less(val interface{}) *Ql {
	s.QlWhere.Less(val)
	return s
}

/**
* MoreEq
* @param val interface{}
* @return *Ql
**/
func (s *Ql) MoreEq(val interface{}) *Ql {
	s.QlWhere.MoreEq(val)
	return s
}

/**
* LessEq
* @param val interface{}
* @return *Ql
**/
func (s *Ql) LessEq(val interface{}) *Ql {
	s.QlWhere.LessEq(val)
	return s
}

/*
*
* Between
* @param vals interface{}
* @return *Ql
**/
func (s *Ql) Between(vals interface{}) *Ql {
	s.QlWhere.Between(vals)
	return s
}

/**
* IsNull
* @return *Ql
**/
func (s *Ql) IsNull() *Ql {
	s.QlWhere.IsNull()
	return s
}

/**
* NotNull
* @return *Ql
**/
func (s *Ql) NotNull() *Ql {
	s.QlWhere.NotNull()
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

/**
* setWheres
* @param wheres et.Json
* @return *Ql
**/
func (s *Ql) setWheres(wheres et.Json) *Ql {
	if s.QlWhere == nil {
		s.QlWhere = newQlWhere()
	}
	s.QlWhere.setWheres(wheres)
	return s
}

/**
* getWhereByPrimaryKeys
* @param data et.Json
* @return error
**/
func (s *Ql) getWhereByPrimaryKeys(data et.Json) error {
	from := s.Froms.Froms[0]
	for name, col := range from.PrimaryKeys {
		val, exists := data[name]
		if !exists {
			return fmt.Errorf("primary key %s is required in model:%s", name, from.Name)
		}
		s.Where(col.Name).Eq(val)
	}

	return nil
}

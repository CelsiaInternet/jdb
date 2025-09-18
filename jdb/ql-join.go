package jdb

import (
	"encoding/json"
	"fmt"

	"github.com/celsiainternet/elvis/console"
	"github.com/celsiainternet/elvis/et"
)

type TypeJoin int

const (
	InnerJoin TypeJoin = iota
	LeftJoin
	RightJoin
	FullJoin
)

type QlJoin struct {
	Ql       *Ql         `json:"-"`
	TypeJoin TypeJoin    `json:"type_join"`
	From     *QlFrom     `json:"from"`
	With     *QlFrom     `json:"with"`
	Field    *Field      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

/**
* QlJoin
* @param name interface{}
* @return *Ql
**/
func (s *Ql) join(tp TypeJoin, from *QlFrom, with *Model, field string, operator string, value interface{}) *Ql {
	if from == nil {
		return s
	}

	result := &QlJoin{
		Ql:       s,
		TypeJoin: tp,
		From:     from,
		With:     s.Froms.add(with),
		Operator: operator,
	}

	result.Field = from.getField(field, false)
	switch v := value.(type) {
	case string:
		result.Value = with.getField(v, false)
	default:
		result.Value = v
	}

	s.Joins = append(s.Joins, result)

	return s
}

/**
* Join
* @param with *Model, field string, operator string, value interface{}
* @return *Ql
**/
func (s *Ql) Join(with *Model, field string, operator string, value interface{}) *Ql {
	var from *QlFrom
	n := len(s.Joins)
	if n == 0 {
		from = s.Froms.getForm(0)
	} else {
		from = s.Joins[n-1].With
	}

	return s.join(InnerJoin, from, with, field, operator, value)
}

/**
* LeftJoin
* @param with *Model, field string, operator string, value interface{}
* @return *Ql
**/
func (s *Ql) LeftJoin(with *Model, field string, operator string, value interface{}) *Ql {
	var from *QlFrom
	n := len(s.Joins)
	if n == 0 {
		from = s.Froms.getForm(0)
	} else {
		from = s.Joins[n-1].With
	}

	return s.join(LeftJoin, from, with, field, operator, value)
}

/**
* RightJoin
* @param with *Model, field string, operator string, value interface{}
* @return *Ql
**/
func (s *Ql) RightJoin(with *Model, field string, operator string, value interface{}) *Ql {
	var from *QlFrom
	n := len(s.Joins)
	if n == 0 {
		from = s.Froms.getForm(0)
	} else {
		from = s.Joins[n-1].With
	}

	return s.join(RightJoin, from, with, field, operator, value)
}

/**
* FullJoin
* @param with *Model, field string, operator string, value interface{}
* @return *Ql
**/
func (s *Ql) FullJoin(with *Model, field string, operator string, value interface{}) *Ql {
	var from *QlFrom
	n := len(s.Joins)
	if n == 0 {
		from = s.Froms.getForm(0)
	} else {
		from = s.Joins[n-1].With
	}

	return s.join(FullJoin, from, with, field, operator, value)
}

/**
* Serialize
* @return []byte, error
**/
func (s *QlJoin) Serialize() ([]byte, error) {
	result, err := json.Marshal(s)
	if err != nil {
		return []byte{}, err
	}

	return result, nil
}

/**
* Describe
* @return *et.Json
**/
func (s *QlJoin) Describe() et.Json {
	definition, err := s.Serialize()
	if err != nil {
		console.Alert(fmt.Sprintf("QlJoin error: %s", err.Error()))
		return et.Json{}
	}

	result := et.Json{}
	err = json.Unmarshal(definition, &result)
	if err != nil {
		console.Alert(fmt.Sprintf("QlJoin error: %s", err.Error()))
		return et.Json{}
	}

	result["ql"] = s.Ql.Describe()

	return result
}

/**
* SetJoins
* @param joins []et.Json
**/
func (s *Ql) setJoins(joins []et.Json) *Ql {
	for _, join := range joins {
		sWith := join.Str("with")
		with := s.Db.GetModel(sWith)
		if with == nil {
			continue
		}

		field := join.Str("field")
		operator := join.Str("operator")
		value := join.Str("value")
		s.Join(with, field, operator, value)
	}

	return s
}

/**
* getJoins
* @return []et.Json
**/
func (s *Ql) getJoins() []et.Json {
	result := []et.Json{}
	for _, join := range s.Joins {
		item := et.Json{
			"with":     join.With.Name,
			"field":    join.Field.Name,
			"operator": join.Operator,
			"value":    join.Value,
		}
		result = append(result, item)
	}

	return result
}

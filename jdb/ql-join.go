package jdb

import (
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
	TypeJoin  TypeJoin       `json:"type_join"`
	From      *QlFrom        `json:"from"`
	With      *QlFrom        `json:"with"`
	Condition []*QlCondition `json:"condition"`
}

/**
* QlJoin
* @param tp TypeJoin, from, with string, conditions []*QlCondition
* @return *Ql
**/
func (s *Ql) join(tp TypeJoin, from, with *QlFrom, conditions []*QlCondition) *Ql {
	result := &QlJoin{
		TypeJoin:  tp,
		From:      from,
		With:      with,
		Condition: conditions,
	}
	s.Joins = append(s.Joins, result)
	return s
}

/**
* Join
* @param withName string, field interface{}, operator Operator, value interface{}
* @return *Ql
**/
func (s *Ql) Join(withName string, fieldFrom string, operator Operator, value interface{}) *Ql {
	with := s.Froms.getFrom(withName)
	if with == nil {
		withForm := s.GetModel(withName)
		if withForm == nil {
			return s
		}
		with = s.Froms.add(withForm)
	}

	field := s.getField(fieldFrom)
	if field == nil {
		return s
	}

	from := field.Model
	if from == nil {
		return s
	}

	condition := &QlCondition{
		Field:    field,
		Operator: operator,
	}
	condition.setValue(value)
	return s.join(InnerJoin, from, with, []*QlCondition{condition})
}

/**
* LeftJoin
* @param withName string, fieldFrom string, operator Operator, value interface{}
* @return *Ql
**/
func (s *Ql) LeftJoin(withName string, fieldFrom string, operator Operator, value interface{}) *Ql {
	with := s.Froms.getFrom(withName)
	if with == nil {
		withForm := s.GetModel(withName)
		if withForm == nil {
			return s
		}
		with = s.Froms.add(withForm)
	}

	field := s.getField(fieldFrom)
	if field == nil {
		return s
	}

	from := field.Model
	if from == nil {
		return s
	}

	condition := &QlCondition{
		Field:    field,
		Operator: operator,
	}
	condition.setValue(value)
	return s.join(LeftJoin, from, with, []*QlCondition{condition})
}

/**
* RightJoin
* @param with *Model, field string, operator string, value interface{}
* @return *Ql
**/
func (s *Ql) RightJoin(withName string, fieldFrom string, operator Operator, value interface{}) *Ql {
	with := s.Froms.getFrom(withName)
	if with == nil {
		withForm := s.GetModel(withName)
		if withForm == nil {
			return s
		}
		with = s.Froms.add(withForm)
	}

	field := s.getField(fieldFrom)
	if field == nil {
		return s
	}

	from := field.Model
	if from == nil {
		return s
	}

	condition := &QlCondition{
		Field:    field,
		Operator: operator,
	}
	condition.setValue(value)
	return s.join(RightJoin, from, with, []*QlCondition{condition})
}

/**
* FullJoin
* @param with *Model, field string, operator string, value interface{}
* @return *Ql
**/
func (s *Ql) FullJoin(withName string, fieldFrom string, operator Operator, value interface{}) *Ql {
	with := s.Froms.getFrom(withName)
	if with == nil {
		withForm := s.GetModel(withName)
		if withForm == nil {
			return s
		}
		with = s.Froms.add(withForm)
	}

	field := s.getField(fieldFrom)
	if field == nil {
		return s
	}

	from := field.Model
	if from == nil {
		return s
	}

	condition := &QlCondition{
		Field:    field,
		Operator: operator,
	}
	condition.setValue(value)
	return s.join(FullJoin, from, with, []*QlCondition{condition})
}

/**
* ToJson
* @return *et.Json
**/
func (s *QlJoin) ToJson() et.Json {
	conditions := []et.Json{}
	for _, condition := range s.Condition {
		conditions = append(conditions, condition.ToJson())
	}
	result := et.Json{
		"type_join": s.TypeJoin,
		"from":      s.From.ToJson(),
		"with":      s.With.ToJson(),
		"condition": conditions,
	}
	return result
}

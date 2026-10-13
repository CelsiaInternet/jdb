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
* @param withName interface{}, field interface{}, operator interface{}, value interface{}
* @return *Ql
**/
func (s *Ql) Join(withName interface{}, fieldFrom string, operator interface{}, value interface{}) *Ql {
	var with *QlFrom
	switch v := withName.(type) {
	case string:
		with = s.Froms.getFrom(v)
	case *QlFrom:
		with = v
	case *Model:
		with = s.Froms.add(v)
	default:
		return s
	}

	if with == nil {
		return s
	}

	field := s.getField(fieldFrom)
	if field == nil {
		return s
	}

	from := field.Model
	if from == nil {
		return s
	}

	var opt Operator
	switch v := operator.(type) {
	case Operator:
		opt = v
	case string:
		opt = StrToOperator(v)
	default:
		return s
	}

	condition := &QlCondition{
		Field:    field,
		Operator: opt,
	}
	condition.setValue(value)
	return s.join(InnerJoin, from, with, []*QlCondition{condition})
}

/**
* LeftJoin
* @param withName interface{}, fieldFrom string, operator interface{}, value interface{}
* @return *Ql
**/
func (s *Ql) LeftJoin(withName interface{}, fieldFrom string, operator interface{}, value interface{}) *Ql {
	var with *QlFrom
	switch v := withName.(type) {
	case string:
		with = s.Froms.getFrom(v)
	case *QlFrom:
		with = v
	case *Model:
		with = s.Froms.add(v)
	default:
		return s
	}

	if with == nil {
		return s
	}

	field := s.getField(fieldFrom)
	if field == nil {
		return s
	}

	from := field.Model
	if from == nil {
		return s
	}

	var opt Operator
	switch v := operator.(type) {
	case Operator:
		opt = v
	case string:
		opt = StrToOperator(v)
	default:
		return s
	}

	condition := &QlCondition{
		Field:    field,
		Operator: opt,
	}
	condition.setValue(value)
	return s.join(LeftJoin, from, with, []*QlCondition{condition})
}

/**
* RightJoin
* @param withName interface{}, field string, operator interface{}, value interface{}
* @return *Ql
**/
func (s *Ql) RightJoin(withName interface{}, fieldFrom string, operator interface{}, value interface{}) *Ql {
	var with *QlFrom
	switch v := withName.(type) {
	case string:
		with = s.Froms.getFrom(v)
	case *QlFrom:
		with = v
	case *Model:
		with = s.Froms.add(v)
	default:
		return s
	}

	if with == nil {
		return s
	}

	field := s.getField(fieldFrom)
	if field == nil {
		return s
	}

	from := field.Model
	if from == nil {
		return s
	}

	var opt Operator
	switch v := operator.(type) {
	case Operator:
		opt = v
	case string:
		opt = StrToOperator(v)
	default:
		return s
	}

	condition := &QlCondition{
		Field:    field,
		Operator: opt,
	}
	condition.setValue(value)
	return s.join(RightJoin, from, with, []*QlCondition{condition})
}

/**
* FullJoin
* @param withName interface{}, field string, operator interface{}, value interface{}
* @return *Ql
**/
func (s *Ql) FullJoin(withName interface{}, fieldFrom string, operator interface{}, value interface{}) *Ql {
	var with *QlFrom
	switch v := withName.(type) {
	case string:
		with = s.Froms.getFrom(v)
	case *QlFrom:
		with = v
	case *Model:
		with = s.Froms.add(v)
	default:
		return s
	}

	if with == nil {
		return s
	}

	field := s.getField(fieldFrom)
	if field == nil {
		return s
	}

	from := field.Model
	if from == nil {
		return s
	}

	var opt Operator
	switch v := operator.(type) {
	case Operator:
		opt = v
	case string:
		opt = StrToOperator(v)
	default:
		return s
	}

	condition := &QlCondition{
		Field:    field,
		Operator: opt,
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

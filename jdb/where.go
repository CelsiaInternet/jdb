package jdb

import (
	"fmt"
	"strings"
	"time"

	"github.com/celsiainternet/elvis/et"
)

type Connector int

const (
	NoC Connector = iota
	And
	Or
)

func (s Connector) Str() string {
	switch s {
	case And:
		return "and"
	case Or:
		return "or"
	default:
		return ""
	}
}

type Operator int

const (
	NoP Operator = iota
	Equal
	Neg
	In
	NotIn
	Like
	More
	Less
	MoreEq
	LessEq
	Between
	IsNull
	NotNull
	Search
)

func (s Operator) command() string {
	switch s {
	case Equal:
		return "="
	case Neg:
		return "!="
	case In:
		return "in"
	case Like:
		return "like"
	case More:
		return ">"
	case Less:
		return "<"
	case MoreEq:
		return ">="
	case LessEq:
		return "<="
	case Between:
		return "between"
	case IsNull:
		return "is null"
	case NotNull:
		return "is not null"
	case Search:
		return "search"
	default:
		return "Any"
	}
}

/**
* Command
* @return string
**/
func (s *Operator) str() string {
	switch *s {
	case Equal:
		return "eq"
	case Neg:
		return "neg"
	case In:
		return "in"
	case Like:
		return "like"
	case More:
		return "more"
	case Less:
		return "less"
	case MoreEq:
		return "moreEq"
	case LessEq:
		return "lessEq"
	case Between:
		return "between"
	case IsNull:
		return "isNull"
	case NotNull:
		return "notNull"
	case Search:
		return "search"
	default:
		return "any"
	}
}

/**
* OperatorToCommand
* @param op Operator
* @return string
**/
func OperatorToCommand(op Operator) string {
	return op.command()
}

/**
* StrToOperator
* @param str string
* @return Operator
**/
func StrToOperator(str string) Operator {
	switch str {
	case "eq":
		return Equal
	case "neg":
		return Neg
	case "in":
		return In
	case "like":
		return Like
	case "more":
		return More
	case "less":
		return Less
	case "moreEq":
		return MoreEq
	case "lessEq":
		return LessEq
	case "between":
		return Between
	case "isNull":
		return IsNull
	case "notNull":
		return NotNull
	case "search":
		return Search
	default:
		return NoP
	}
}

type TypeAgregation int

const (
	Nag TypeAgregation = iota
	AgregationSum
	AgregationCount
	AgregationAvg
	AgregationMin
	AgregationMax
	AgregationValue
)

func (s TypeAgregation) Str() string {
	switch s {
	case AgregationSum:
		return "SUM"
	case AgregationCount:
		return "COUNT"
	case AgregationAvg:
		return "AVG"
	case AgregationMin:
		return "MIN"
	case AgregationMax:
		return "MAX"
	case AgregationValue:
		return "VALUE"
	default:
		return ""
	}
}

type Agregation struct {
	Agregation TypeAgregation
	Value      interface{}
}

func (s *Agregation) Str() string {
	return fmt.Sprintf("%s(%v)", s.Agregation.Str(), s.Value)
}

func SUM(value interface{}) *Agregation {
	return &Agregation{
		Agregation: AgregationSum,
		Value:      value,
	}
}

func COUNT(value interface{}) *Agregation {
	return &Agregation{
		Agregation: AgregationCount,
		Value:      value,
	}
}

func AVG(value interface{}) *Agregation {
	return &Agregation{
		Agregation: AgregationAvg,
		Value:      value,
	}
}

func MIN(value interface{}) *Agregation {
	return &Agregation{
		Agregation: AgregationMin,
		Value:      value,
	}
}

func MAX(value interface{}) *Agregation {
	return &Agregation{
		Agregation: AgregationMax,
		Value:      value,
	}
}

func VALUE(value interface{}) *Agregation {
	return &Agregation{
		Agregation: AgregationValue,
		Value:      value,
	}
}

type ValueType int

const (
	ValueTypeString ValueType = iota
	ValueTypeNumber
	ValueTypeDateTime
	ValueTypeBoolean
	ValueTypeJson
	ValueTypeJsonArray
	ValueTypeArray
	ValueTypeBinary
	ValueTypeNull
	// Calc
	ValueTypeAgregation
	ValueTypeField
	ValueTypeCalc
)

type Value struct {
	Type  ValueType `json:"type"`
	Value any       `json:"value"`
}

type QlCondition struct {
	Connector Connector   `json:"connector"`
	Field     interface{} `json:"field"`
	Operator  Operator    `json:"operator"`
	Value     Value       `json:"value"`
}

/**
* newQlCondition
* @params field interface{}
* @return QlWhere
**/
func newQlCondition(field interface{}) *QlCondition {
	result := &QlCondition{
		Connector: NoC,
		Field:     field,
		Operator:  NoP,
	}
	result.setValue("")
	return result
}

func (s *QlCondition) fieldToString() string {
	switch v := s.Field.(type) {
	case *Field:
		return v.asName()
	case Field:
		return v.asName()
	case *Agregation:
		return v.Str()
	default:
		return fmt.Sprintf(`%v`, Quote(v))
	}
}

func (s *QlCondition) ToJson() et.Json {
	return et.Json{
		s.fieldToString(): et.Json{
			s.Operator.str(): s.Value,
		},
	}
}

func (s *QlCondition) setValue(value interface{}) {
	switch v := value.(type) {
	case *Field:
		s.Value = Value{
			Type:  ValueTypeField,
			Value: v,
		}
	case Field:
		s.Value = Value{
			Type:  ValueTypeField,
			Value: v,
		}
	case *Agregation:
		s.Value = Value{
			Type:  ValueTypeAgregation,
			Value: v,
		}
	case Agregation:
		s.Value = Value{
			Type:  ValueTypeAgregation,
			Value: v,
		}
	case string:
		s.Value = Value{
			Type:  ValueTypeString,
			Value: v,
		}
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		s.Value = Value{
			Type:  ValueTypeNumber,
			Value: v,
		}
	case float32, float64:
		s.Value = Value{
			Type:  ValueTypeNumber,
			Value: v,
		}
	case bool:
		s.Value = Value{
			Type:  ValueTypeBoolean,
			Value: v,
		}
	case time.Time:
		s.Value = Value{
			Type:  ValueTypeDateTime,
			Value: v,
		}
	case et.Json, map[string]interface{}:
		s.Value = Value{
			Type:  ValueTypeJson,
			Value: v,
		}
	case []et.Json, []map[string]interface{}:
		s.Value = Value{
			Type:  ValueTypeJsonArray,
			Value: v,
		}
	case []interface{}, []string, []int, []int8, []int16, []int32, []int64, []uint, []uint16, []uint32, []uint64, []float32, []float64:
		s.Value = Value{
			Type:  ValueTypeArray,
			Value: v,
		}
	case []byte:
		s.Value = Value{
			Type:  ValueTypeBinary,
			Value: v,
		}
	case nil:
		s.Value = Value{
			Type:  ValueTypeNull,
			Value: nil,
		}
	default:
		s.Value = Value{
			Type:  ValueTypeString,
			Value: fmt.Sprintf(`%v`, v),
		}
	}
}

type QlWhere struct {
	Wheres  []*QlCondition `json:"wheres"`
	IsDebug bool           `json:"-"`
}

/**
* newQlWhere
* @return *QlWhere
**/
func newQlWhere() *QlWhere {
	return &QlWhere{
		Wheres:  []*QlCondition{},
		IsDebug: false,
	}
}

/**
* Debug
* @return *QlWhere
**/
func (s *QlWhere) Debug() *QlWhere {
	s.IsDebug = true

	return s
}

/**
* addCondition
* @param condition *QlCondition
* @return *QlWhere
**/
func (s *QlWhere) addCondition(condition *QlCondition) *QlWhere {
	s.Wheres = append(s.Wheres, condition)
	return s
}

/**
* setWhere
* @param val field interface{}
* @return *QlWhere
**/
func (s *QlWhere) setWhere(field interface{}) *QlWhere {
	where := newQlCondition(field)
	if len(s.Wheres) > 0 {
		where.Connector = And
	}

	return s.addCondition(where)
}

/**
* setOr
* @param val field interface{}
* @return *QlWhere
**/
func (s *QlWhere) setOr(field interface{}) *QlWhere {
	where := newQlCondition(field)
	where.Connector = Or
	return s.addCondition(where)
}

/**
* condition
* @return *QlCondition
**/
func (s *QlWhere) condition() *QlCondition {
	idx := len(s.Wheres)
	if idx <= 0 {
		return nil
	}

	return s.Wheres[idx-1]
}

/**
* Where
* @param fld interface{}
* @return *QlWhere
**/
func (s *QlWhere) Where(fld interface{}) *QlWhere {
	return s.setWhere(fld)
}

/**
* And
* @param fld interface{}
* @return *QlWhere
**/
func (s *QlWhere) And(fld interface{}) *QlWhere {
	return s.Where(fld)
}

/**
* Or
* @param val interface{}
* @return *QlWhere
**/
func (s *QlWhere) Or(fld interface{}) *QlWhere {
	return s.setOr(fld)
}

/**
* Eq
* @param val interface{}
* @return QlWhere
**/
func (s *QlWhere) Eq(val interface{}) *QlWhere {
	condition := s.condition()
	if condition == nil {
		return s
	}

	condition.Operator = Equal
	condition.setValue(val)
	return s
}

/**
* Neg
* @param val interface{}
* @return QlWhere
**/
func (s *QlWhere) Neg(val interface{}) *QlWhere {
	condition := s.condition()
	if condition == nil {
		return s
	}

	condition.Operator = Neg
	condition.setValue(val)
	return s
}

/**
* In
* @param val ...any
* @return QlWhere
**/
func (s *QlWhere) In(val ...any) *QlWhere {
	condition := s.condition()
	if condition == nil {
		return s
	}

	condition.Operator = In
	condition.setValue(val)
	return s
}

/**
* NotIn
* @param val ...any
* @return QlWhere
**/
func (s *QlWhere) NotIn(val ...any) *QlWhere {
	condition := s.condition()
	if condition == nil {
		return s
	}

	condition.Operator = NotIn
	condition.setValue(val)
	return s
}

/**
* Like
* @param val interface{}
* @return QlWhere
**/
func (s *QlWhere) Like(val interface{}) *QlWhere {
	condition := s.condition()
	if condition == nil {
		return s
	}

	condition.Operator = Like
	condition.setValue(val)
	return s
}

/**
* More
* @param val interface{}
* @return QlWhere
**/
func (s *QlWhere) More(val interface{}) *QlWhere {
	condition := s.condition()
	if condition == nil {
		return s
	}

	condition.Operator = More
	condition.setValue(val)
	return s
}

/**
* Less
* @param val interface{}
* @return QlWhere
**/
func (s *QlWhere) Less(val interface{}) *QlWhere {
	condition := s.condition()
	if condition == nil {
		return s
	}

	condition.Operator = Less
	condition.setValue(val)
	return s
}

/**
* MoreEq
* @param val interface{}
* @return QlWhere
**/
func (s *QlWhere) MoreEq(val interface{}) *QlWhere {
	condition := s.condition()
	if condition == nil {
		return s
	}

	condition.Operator = MoreEq
	condition.setValue(val)
	return s
}

/**
* LessEq
* @param val interface{}
* @return QlWhere
**/
func (s *QlWhere) LessEq(val interface{}) *QlWhere {
	condition := s.condition()
	if condition == nil {
		return s
	}

	condition.Operator = LessEq
	condition.setValue(val)
	return s
}

/**
* Between
* @param val1, val2 interface{}
* @return QlWhere
**/
func (s *QlWhere) Between(vals interface{}) *QlWhere {
	val, ok := vals.([]interface{})
	if !ok {
		return s
	}

	condition := s.condition()
	if condition == nil {
		return s
	}

	condition.Operator = Between
	condition.setValue(val)
	return s
}

/**
* IsNull
* @return *QlWhere
**/
func (s *QlWhere) IsNull() *QlWhere {
	condition := s.condition()
	if condition == nil {
		return s
	}

	condition.Operator = IsNull
	return s
}

/**
* NotNull
* @return *QlWhere
**/
func (s *QlWhere) NotNull() *QlWhere {
	condition := s.condition()
	if condition == nil {
		return s
	}

	condition.Operator = NotNull
	return s
}

/**
* getWheres
* @return et.Json
**/
func (s *QlWhere) getWheres() et.Json {
	result := et.Json{}
	and := []et.Json{}
	or := []et.Json{}
	for i, condition := range s.Wheres {
		if condition.Field == nil {
			continue
		}

		if condition.Connector == And {
			and = append(and, condition.ToJson())
		} else if condition.Connector == Or {
			or = append(or, condition.ToJson())
		} else if i == 0 {
			result = condition.ToJson()
		}
	}

	if len(and) > 0 {
		result.Set("AND", and)
	}
	if len(or) > 0 {
		result.Set("OR", or)
	}

	return result
}

/**
* setWheres
* @param wheres et.Json
* @return *QlWhere
**/
func (s *QlWhere) setWheres(wheres et.Json) *QlWhere {
	setOperatorValue := func(value et.Json) {
		for key, val := range value {
			switch key {
			case "eq":
				s.Eq(val)
			case "neg":
				s.Neg(val)
			case "in":
				s.In(val)
			case "like":
				s.Like(val)
			case "more":
				s.More(val)
			case "less":
				s.Less(val)
			case "moreEq":
				s.MoreEq(val)
			case "lessEq":
				s.LessEq(val)
			case "between":
				s.Between(val)
			case "isNull":
				s.IsNull()
			case "notNull":
				s.NotNull()
			}
		}
	}

	for key := range wheres {
		if strings.ToLower(key) == "AND" {
			value := wheres.ArrayJson(key)
			for _, where := range value {
				for key := range where {
					val := wheres.Json(key)
					s.setWhere(key)
					setOperatorValue(val)
				}
			}
		} else if strings.ToLower(key) == "OR" {
			value := wheres.ArrayJson(key)
			for _, where := range value {
				for key := range where {
					val := wheres.Json(key)
					s.setOr(key)
					setOperatorValue(val)
				}
			}
		} else {
			val := wheres.Json(key)
			s.setWhere(key)
			setOperatorValue(val)
		}
	}

	return s
}

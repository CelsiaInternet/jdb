package jdb

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/strs"
)

type TypeAgregation int

const (
	Nag TypeAgregation = iota
	AgregationSum
	AgregationCount
	AgregationAvg
	AgregationMin
	AgregationMax
	ExtractYear
	ExtractMonth
	ExtractDay
	ExtractHour
	ExtractMinute
	ExtractSecond
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
	case ExtractYear:
		return "YEAR"
	case ExtractMonth:
		return "MONTH"
	case ExtractDay:
		return "DAY"
	case ExtractHour:
		return "HOUR"
	case ExtractMinute:
		return "MINUTE"
	case ExtractSecond:
		return "SECOND"
	}

	return ""
}

type TypeResult int

const (
	TpResult TypeResult = iota
	TpList
)

/**
* StrToTypeResult
* @param str string
* @return TypeResult
**/
func StrToTypeResult(str string) TypeResult {
	switch str {
	case "list":
		return TpList
	}

	return TpResult
}

type Field struct {
	Column     *Column        `json:"-"`
	Schema     string         `json:"schema"`
	Model      string         `json:"model"`
	As         string         `json:"as"`
	Name       string         `json:"name"`
	Source     string         `json:"source"`
	Agregation TypeAgregation `json:"agregation"`
	Value      interface{}    `json:"value"`
	Alias      string         `json:"alias"`
	Hidden     bool           `json:"hidden"`
	Page       int            `json:"page"`
	Rows       int            `json:"rows"`
	TpResult   TypeResult     `json:"tp_result"`
	Unquoted   bool           `json:"unquoted"`
	Select     []interface{}  `json:"select"`
	Joins      []et.Json      `json:"joins"`
	Where      et.Json        `json:"where"`
	GroupBy    []string       `json:"group_by"`
	Havings    et.Json        `json:"havings"`
	OrderBy    et.Json        `json:"order_by"`
}

func (s *Field) describe() et.Json {
	return et.Json{
		"column_type": s.Column.TypeColumn.Str(),
		"schema":      s.Schema,
		"model":       s.Model,
		"as":          s.As,
		"name":        s.Name,
		"source":      s.Source,
		"agregation":  s.Agregation.Str(),
		"value":       s.Value,
		"alias":       s.Alias,
		"hidden":      s.Hidden,
		"page":        s.Page,
		"rows":        s.Rows,
		"tp_result":   s.TpResult,
		"unquoted":    s.Unquoted,
		"select":      s.Select,
		"joins":       s.Joins,
		"where":       s.Where,
		"group_by":    s.GroupBy,
		"havings":     s.Havings,
		"order_by":    s.OrderBy,
	}
}

/**
* newField
* @param name string
* @return *Field
**/
func newField(name string) *Field {
	return &Field{
		Name:    name,
		Select:  make([]interface{}, 0),
		Joins:   make([]et.Json, 0),
		Where:   et.Json{},
		GroupBy: make([]string, 0),
		Havings: et.Json{},
		OrderBy: et.Json{},
	}
}

/**
* Serialize
* @return []byte, error
**/
func (s *Field) Serialize() ([]byte, error) {
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
func (s *Field) Describe() et.Json {
	definition, err := s.Serialize()
	if err != nil {
		return et.Json{}
	}

	result := et.Json{}
	err = json.Unmarshal(definition, &result)
	if err != nil {
		return et.Json{}
	}

	result["column"] = s.Column.Describe()

	return result
}

/**
* setValue
* @param value interface{}
**/
func (s *Field) setValue(value interface{}) {
	regexpMust := func(pattern string, value interface{}) (string, bool) {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(value.(string))

		if len(matches) > 1 {
			return matches[1], true
		} else {
			return value.(string), false
		}
	}

	switch v := value.(type) {
	case string:
		result, ok := regexpMust(`(?i)^CALC\((.*)\)$`, v)
		if ok {
			s.Value = result
			s.Unquoted = true
		} else {
			re := regexp.MustCompile(`^:(.*)`)
			matches := re.FindStringSubmatch(v)
			if len(matches) > 1 {
				s.Value = matches[1]
				s.Unquoted = true
			} else {
				s.Value = v
			}
		}
	default:
		s.Value = value
	}
}

/**
* setAgregation
* @param agr TypeAgregation
**/
func (s *Field) setAgregation(agr TypeAgregation) {
	s.Agregation = agr
	switch agr {
	case AgregationSum:
		s.Alias = strs.Format("sum_%s", s.Name)
	case AgregationCount:
		s.Alias = strs.Format("count_%s", s.Name)
	case AgregationAvg:
		s.Alias = strs.Format("avg_%s", s.Name)
	case AgregationMin:
		s.Alias = strs.Format("min_%s", s.Name)
	case AgregationMax:
		s.Alias = strs.Format("max_%s", s.Name)
	}
}

/**
* ValueArg
* @return string
**/
func (s *Field) ValueArg() string {
	switch v := s.Value.(type) {
	case time.Time:
		val := v.Format(time.RFC3339)
		return fmt.Sprintf(`%v`, val)
	default:
		if s.Value == nil {
			return fmt.Sprintf(`'%v'`, "NULL")
		}
		if s.Value == "nil" {
			return fmt.Sprintf(`'%v'`, "NULL")
		}
		return strs.Format(`%v`, s.Value)
	}
}

/**
* ValueQuoted
* @return any
**/
func (s *Field) ValueQuoted() interface{} {
	if s.Unquoted {
		return s.Value
	}

	if s.Column != nil && s.Column.TypeData == TypeDataDateTime {
		f := "2006/01/02 15:04:05"
		v := fmt.Sprintf(`%v`, s.Value)
		t, err := time.Parse(f, v)
		if err != nil {
			f = "2006-01-02 15:04:05"
			t, err = time.Parse(f, v)
			if err != nil {
				return v
			}
		}

		return Quote(t)
	}

	return Quote(s.Value)
}

/**
* ValueToJSON
* @return any
**/
func (s *Field) ValueToJSON() (any, string) {
	switch v := s.Value.(type) {
	case string:
		val := EscapeJSON(v)
		return fmt.Sprintf(`'%v'`, val), "text"
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return v, "numeric"
	case float32, float64:
		return v, "numeric"
	case bool:
		return v, "boolean"
	case time.Time:
		val := v.Format(time.RFC3339)
		return fmt.Sprintf(`'%v'`, val), "timestamptz"
	case et.Json:
		val := EscapeJSON(v.ToString())
		return fmt.Sprintf(`'%v'`, val), "jsonb"
	case et.Items:
		val := EscapeJSON(v.ToString())
		return fmt.Sprintf(`'%v'`, val), "jsonb"
	case et.Item:
		val := EscapeJSON(v.ToString())
		return fmt.Sprintf(`'%v'`, val), "jsonb"
	case map[string]interface{}:
		result, err := json.Marshal(s.Value)
		if err != nil {
			return fmt.Sprintf(`'%v'`, "{}"), "jsonb"
		}
		val := EscapeJSON(string(result))
		return fmt.Sprintf(`'%v'`, val), "jsonb"
	case []map[string]interface{}:
		result, err := json.Marshal(s.Value)
		if err != nil {
			return fmt.Sprintf(`'%v'`, "[]"), "jsonb"
		}
		val := EscapeJSON(string(result))
		return fmt.Sprintf(`'%v'`, val), "jsonb"
	case []et.Json:
		result, err := json.Marshal(s.Value)
		if err != nil {
			return fmt.Sprintf(`'%v'`, "[]"), "jsonb"
		}
		val := EscapeJSON(string(result))
		return fmt.Sprintf(`'%v'`, val), "jsonb"
	default:
		if v == nil {
			return fmt.Sprintf(`'%v'`, "null"), "jsonb"
		}
		return strs.Format(`%v`, v), "text"
	}
}

/**
* asField
* @return string
**/
func (s *Field) asField() string {
	result := ""
	result = strs.Append(result, s.Schema, "")
	result = strs.Append(result, s.Model, ".")
	result = strs.Append(result, s.Source, ".")
	result = strs.Append(result, s.Name, ".")

	return result
}

/**
* asName
* @return string
**/
func (s *Field) asName() string {
	if s.As != "" {
		return strs.Format(`%s.%s`, s.As, s.Name)
	}

	return strs.Format(`%v`, s.Name)
}

/**
* GetField
* @return *Field
**/
func GetField(col *Column) *Field {
	result := newField(col.Name)
	result.Column = col
	result.Name = col.Name
	result.Alias = col.Name
	result.Hidden = col.Hidden

	if col.TypeColumn == TpRelatedTo {
		result.Page = 1
		result.Rows = 30
		result.TpResult = TpResult
	}

	if col.Model == nil {
		return result
	}

	result.Schema = col.Model.Schema
	result.Model = col.Model.Name

	if col.Source == nil {
		return result
	}

	result.Source = col.Source.Name

	return result
}

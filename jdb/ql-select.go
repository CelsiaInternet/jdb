package jdb

import (
	"slices"

	"github.com/celsiainternet/elvis/et"
)

/**
* getField
* @param name string
* @return *Field
**/
func (s *Ql) getField(name string) *Field {
	return s.Froms.getField(name)
}

/**
* setSelect
* @param field *Field
* @return *Ql
**/
func (s *Ql) setSelectField(field *Field) *Ql {
	if field == nil {
		return s
	}

	if slices.Contains([]TypeColumn{TpColumn, TpAtribute}, field.TypeColumn) {
		idx := slices.IndexFunc(s.Selects, func(e *Field) bool { return e == field })
		if idx == -1 {
			s.Selects = append(s.Selects, field)
		}
	} else if field.TypeColumn == TpRollup {
		if field.Model == nil || field.Model.Model == nil {
			return s
		}

		rollup, exist := field.Model.Model.Rollup[field.Name]
		if !exist {
			return s
		}

		s.Rollups = append(s.Rollups, rollup)
	} else if field.TypeColumn == TpRelatedTo {
		if field.Model == nil || field.Model.Model == nil {
			return s
		}

		relation, exist := field.Model.Model.RelationsTo[field.Name]
		if !exist {
			return s
		}

		s.Details[field.Name] = relation
	} else if field.TypeColumn == TpDetail {
		if field.Model == nil || field.Model.Model == nil {
			return s
		}

		detail, exist := field.Model.Model.Detail[field.Name]
		if !exist {
			return s
		}

		s.Details[field.Name] = detail
	} else if field.TypeColumn == TpCalc {
		if field.Model == nil || field.Model.Model == nil {
			return s
		}

		calc, exist := field.Model.Model.CalcFunction[field.Name]
		if !exist {
			return s
		}

		s.CalcFunction[field.Name] = calc
	}

	return s
}

/**
* Select
* @param fields ...interface{}
* @return *Ql
**/
func (s *Ql) Select(fields ...interface{}) *Ql {
	setRelationTo := func(v map[string]interface{}) {
		for key := range v {
			field := s.getField(key)
			if field != nil {
				s.setSelectField(field)
			}
		}
	}

	for _, name := range fields {
		switch v := name.(type) {
		case string:
			field := s.getField(v)
			s.setSelectField(field)
		case *Column:
			field := s.getField(v.Name)
			s.setSelectField(field)
		case Column:
			field := s.getField(v.Name)
			s.setSelectField(field)
		case *Field:
			s.setSelectField(v)
		case *Agregation:
			s.setSelectAgregation(v)
		case Agregation:
			s.setSelectAgregation(v)
		case et.Json:
			setRelationTo(v)
		case map[string]interface{}:
			setRelationTo(v)
		}
	}

	s.TypeSelect = Select
	return s
}

/**
* Data
* @param fields ...interface{}
* @return *Ql
**/
func (s *Ql) Data(fields ...interface{}) *Ql {
	result := s.Select(fields...)
	result.TypeSelect = Source
	return result
}

/**
* Detail
* @param fields ...interface{}
* @return *Ql
**/
func (s *Ql) Detail(fields ...interface{}) *Ql {
	setDetail := func(name string) {
		field := s.getField(name)
		if map[TypeColumn]bool{TpRelatedTo: true, TpCalc: true, TpRollup: true}[field.TypeColumn] {
			s.setSelect(field)
		}
	}

	for _, name := range fields {
		switch v := name.(type) {
		case string:
			setDetail(v)
		case *Column:
			setDetail(v.Name)
		}
	}

	return s
}

/**
* Hidden
* @param fields ...string
* @return *Ql
**/
func (s *Ql) Hidden(fields ...string) *Ql {
	return s.setHidden(fields...)
}

/**
* setSelects
* @param fields ...interface{}
* @return *Ql
**/
func (s *Ql) setSelects(fields ...interface{}) *Ql {
	froms := s.Froms.Froms
	if len(froms) == 0 {
		return s
	}

	model := froms[0].Model
	if model == nil {
		return s
	}

	if model.SourceField != nil {
		s.Data(fields...)
	} else {
		s.Select(fields...)
	}

	return s
}

/**
* setHidden
* @param columns ...*Column
* @return *Ql
**/
func (s *Ql) setHidden(columns ...string) *Ql {
	s.Hiddens = append(s.Hiddens, columns...)

	return s
}

/**
* getSelects
* @return []string
**/
func (s *Ql) getSelects() []string {
	result := []string{}
	for _, sel := range s.Selects {
		result = append(result, sel.asName())
	}

	return result
}

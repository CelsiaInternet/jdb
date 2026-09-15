package jdb

import (
	"fmt"

	"github.com/celsiainternet/elvis/et"
)

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
* setHavings
* @param havings et.Json
* @return *Ql
**/
func (s *Ql) setHavings(havings et.Json) *Ql {
	if s.Havings == nil {
		s.Havings = newQlWhere()
	}
	s.Havings.setWhere(havings)
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

/**
* Where
* @param fld interface{}
* @return *Ql
**/
func (s *Ql) Where(fld interface{}) *Ql {
	if s.QlWhere == nil {
		s.QlWhere = newQlWhere()
	}
	s.QlWhere.Where(resolveWhereField(fld, s.getField))
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
	s.QlWhere.Or(resolveWhereField(fld, s.getField))
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
* Having
* @param field string
* @return *QlWhere
**/
func (s *Ql) Having(val interface{}) *QlWhere {
	s.Havings.Where(val)
	return s.Havings
}

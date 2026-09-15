package jdb

import "github.com/celsiainternet/elvis/et"

/**
* Where
* @param field string
* @return *Command
**/
func (s *Command) Where(val string) *Command {
	s.QlWhere.Where(resolveWhereField(val, s.getField))

	return s
}

/**
* And
* @param val string
* @return *Command
**/
func (s *Command) And(val string) *Command {
	s.QlWhere.And(resolveWhereField(val, s.getField))

	return s
}

/**
* And
* @param fval string
* @return *Command
**/
func (s *Command) Or(val string) *Command {
	s.QlWhere.Or(resolveWhereField(val, s.getField))

	return s
}

/**
* Eq
* @param val interface{}
* @return *Command
**/
func (s *Command) Eq(val interface{}) *Command {
	s.QlWhere.Eq(val)

	return s
}

/**
* Neg
* @param val interface{}
* @return *Command
**/
func (s *Command) Neg(val interface{}) *Command {
	s.QlWhere.Neg(val)

	return s
}

/**
* In
* @param val ...any
* @return *Command
**/
func (s *Command) In(val ...any) *Command {
	s.QlWhere.In(val...)

	return s
}

/**
* NotIn
* @param val ...any
* @return *Command
**/
func (s *Command) NotIn(val ...any) *Command {
	s.QlWhere.NotIn(val...)

	return s
}

/**
* Like
* @param val interface{}
* @return *Command
**/
func (s *Command) Like(val interface{}) *Command {
	s.QlWhere.Like(val)

	return s
}

/**
* More
* @param val interface{}
* @return *Command
**/
func (s *Command) More(val interface{}) *Command {
	s.QlWhere.More(val)

	return s
}

/**
* Less
* @param val interface{}
* @return *Command
**/
func (s *Command) Less(val interface{}) *Command {
	s.QlWhere.Less(val)

	return s
}

/**
* MoreEq
* @param val interface{}
* @return *Command
**/
func (s *Command) MoreEq(val interface{}) *Command {
	s.QlWhere.MoreEq(val)

	return s
}

/**
* LessEq
* @param val interface{}
* @return *Command
**/
func (s *Command) LessEq(val interface{}) *Command {
	s.QlWhere.LessEq(val)

	return s
}

/*
*
* Between
* @param vals interface{}
* @return *Command
**/
func (s *Command) Between(vals interface{}) *Command {
	s.QlWhere.Between(vals)

	return s
}

/**
* IsNull
* @return *Command
**/
func (s *Command) IsNull() *Command {
	s.QlWhere.IsNull()

	return s
}

/**
* NotNull
* @return *Command
**/
func (s *Command) NotNull() *Command {
	s.QlWhere.NotNull()

	return s
}

/**
* setWhere
* @param setWheres et.Json
* @return *Command
**/
func (s *Command) setWheres(wheres et.Json) *Command {
	if s.QlWhere == nil {
		s.QlWhere = newQlWhere()
	}
	s.QlWhere.setWheres(wheres)
	return s
}

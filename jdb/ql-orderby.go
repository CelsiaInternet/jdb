package jdb

import (
	"github.com/celsiainternet/elvis/et"
)

type QlOrder struct {
	Asc  []*Field
	Desc []*Field
}

/**
* orderBy
* @param asc bool, columns ...string
* @return *Ql
**/
func (s *Ql) orderBy(asc bool, fields ...string) *Ql {
	for _, name := range fields {
		field := s.getField(name)
		if field != nil {
			if asc {
				s.Orders.Asc = append(s.Orders.Asc, field)
			} else {
				s.Orders.Desc = append(s.Orders.Desc, field)
			}
		}
	}

	return s
}

/**
* OrderByAsc
* @param columns ...any
* @return *Ql
**/
func (s *Ql) OrderByAsc(fields ...string) *Ql {
	return s.orderBy(true, fields...)
}

/**
* OrderByDesc
* @param columns ...any
* @return *Ql
**/
func (s *Ql) OrderByDesc(fields ...string) *Ql {
	return s.orderBy(false, fields...)
}

/**
* OrderBy
* @param columns ...any
* @return *Ql
**/
func (s *Ql) OrderBy(fields ...string) *Ql {
	return s.OrderByAsc(fields...)
}

/**
* setOrderBy
* @param orders et.Json
* @return *Ql
**/
func (s *Ql) setOrderBy(orders et.Json) *Ql {
	if len(orders) == 0 {
		return s
	}

	for key := range orders {
		switch key {
		case "asc", "ASC":
			val := orders.ArrayStr(key)
			s.OrderByAsc(val...)
		case "desc", "DESC":
			val := orders.ArrayStr(key)
			s.OrderByDesc(val...)
		}
	}

	return s
}

/**
* getOrders
* @return []string
**/
func (s *Ql) getOrders() et.Json {
	asc := []string{}
	desc := []string{}
	for _, sel := range s.Orders.Asc {
		asc = append(asc, sel.asName())
	}
	for _, sel := range s.Orders.Desc {
		desc = append(desc, sel.asName())
	}

	return et.Json{
		"asc":  asc,
		"desc": desc,
	}
}

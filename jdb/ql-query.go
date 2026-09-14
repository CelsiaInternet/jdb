package jdb

import (
	"fmt"

	"github.com/celsiainternet/elvis/console"
	"github.com/celsiainternet/elvis/et"
)

/**
* FirstTx
* @param tx *Tx, n int
* @return et.Items, error
**/
func (s *Ql) FirstTx(tx *Tx, n int) (et.Items, error) {
	if s.Db == nil {
		return et.Items{}, fmt.Errorf(MSG_DATABASE_IS_REQUIRED)
	}

	s.setTx(tx)
	s.Limit = n
	s.prepare()
	result, err := s.Db.Select(s)
	if err != nil {
		return et.Items{}, err
	}

	for _, data := range result.Result {
		s.GetDetailsTx(tx, data)
	}

	return result, nil
}

/**
* AllTx
* @param tx *Tx
* @return et.Items, error
**/
func (s *Ql) AllTx(tx *Tx) (et.Items, error) {
	return s.FirstTx(tx, 0)
}

/**
* LastTx
* @param tx *Tx, n int
* @return et.Items, error
**/
func (s *Ql) LastTx(tx *Tx, n int) (et.Items, error) {
	return s.FirstTx(tx, n*-1)
}

/**
* OneTx
* @param tx *Tx
* @return et.Item, error
**/
func (s *Ql) OneTx(tx *Tx) (et.Item, error) {
	result, err := s.FirstTx(tx, 1)
	if err != nil {
		return et.Item{}, err
	}

	return result.First(), nil
}

/**
* RowsTx
* @param tx *Tx, limit int
* @return et.Items, error
**/
func (s *Ql) RowsTx(tx *Tx, val int) (et.Items, error) {
	return s.FirstTx(tx, val)
}

/**
* ItExistsTx
* @param tx *Tx
* @return bool, error
**/
func (s *Ql) ItExistsTx(tx *Tx) (bool, error) {
	if s.Db == nil {
		return false, fmt.Errorf(MSG_DATABASE_IS_REQUIRED)
	}

	s.setTx(tx)
	s.prepare()
	result, err := s.Db.Exists(s)
	if err != nil {
		return false, err
	}

	return result, nil
}

/**
* CountedTx
* @param tx *Tx
* @return int, error
**/
func (s *Ql) CountedTx(tx *Tx) (int, error) {
	if s.Db == nil {
		return 0, fmt.Errorf(MSG_DATABASE_IS_REQUIRED)
	}

	s.setTx(tx)
	s.prepare()
	result, err := s.Db.Count(s)
	if err != nil {
		return 0, err
	}

	return result, nil
}

/**
* First
* @param n int
* @return et.Items, error
**/
func (s *Ql) First(n int) (et.Items, error) {
	return s.FirstTx(nil, n)
}

/**
* All
* @return et.Items, error
**/
func (s *Ql) All() (et.Items, error) {
	return s.AllTx(nil)
}

/**
* Last
* @param n int
* @return et.Items, error
**/
func (s *Ql) Last(n int) (et.Items, error) {
	return s.LastTx(nil, n)
}

/**
* One
* @return et.Item, error
**/
func (s *Ql) One() (et.Item, error) {
	return s.OneTx(nil)
}

/**
* Rows
* @param n int
* @return et.Items, error
**/
func (s *Ql) Rows(n int) (et.Items, error) {
	return s.RowsTx(nil, n)
}

/**
* ItExists
* @return bool, error
**/
func (s *Ql) ItExists() (bool, error) {
	return s.ItExistsTx(nil)
}

/**
* Counted
* @return int, error
**/
func (s *Ql) Counted() (int, error) {
	return s.CountedTx(nil)
}

/**
* QueryTx
* @param tx *Tx, params et.Json
* @return et.Json, error
**/
func (s *Ql) QueryTx(tx *Tx, params et.Json) (et.Json, error) {
	return s.queryTx(tx, params)
}

/**
* Query
* @param params et.Json
* @return et.Json, error
**/
func (s *Ql) Query(params et.Json) (et.Json, error) {
	return s.QueryTx(nil, params)
}

/**
* setJoins
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
* SetPage
* @param page int
* @return *Ql
**/
func (s *Ql) setPage(page int) *Ql {
	s.Page(page)

	return s
}

/**
* SetLimitTx
* @param tx *Tx, limit int
* @return et.Json, error
**/
func (s *Ql) setLimitTx(tx *Tx, limit int) (et.Json, error) {
	s.Limit = limit
	if s.Limit <= 0 {
		result, err := s.AllTx(tx)
		if err != nil {
			return nil, err
		}

		res := result.ToJson()
		if s.IsDebug {
			res["sql"] = s.Sql
		}

		return res, nil
	} else if s.Limit == 1 {
		result, err := s.OneTx(tx)
		if err != nil {
			return nil, err
		}

		res := result.ToJson()
		if s.IsDebug {
			res["sql"] = s.Sql
		}

		return res, nil
	} else {
		result, err := s.FirstTx(tx, s.Limit)
		if err != nil {
			return nil, err
		}

		res := result.ToJson()
		if s.IsDebug {
			res["sql"] = s.Sql
		}

		return res, nil
	}
}

/**
* queryTx
* @param tx *Tx, params et.Json
* @return et.Items, error
**/
func (s *Ql) queryTx(tx *Tx, query et.Json) (et.Json, error) {
	selects := query.Array("select")
	joins := query.ArrayJson("join")
	where := query.Json("where")
	groups := query.ArrayStr("group_by")
	havings := query.Json("having")
	orderBy := query.Json("order_by")
	page := query.Int("page")
	limit := query.ValInt(1000, "limit")
	debug := query.Bool("debug")

	if debug {
		console.Debug(query.ToEscapeHTML())
	}

	result, err := s.
		setJoins(joins).
		setWheres(where).
		setGroupBy(groups...).
		setHavings(havings).
		setOrderBy(orderBy).
		setSelects(selects...).
		setDebug(debug).
		setPage(page).
		setLimitTx(tx, limit)
	if err != nil {
		return et.Json{}, err
	}

	return result, nil
}

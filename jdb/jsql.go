package jdb

import (
	"fmt"
	"strings"

	"github.com/celsiainternet/elvis/et"
)

const (
	QUERY_DEFAULT_LIMIT = 100
	QUERY_MAX_LIMIT     = 1000
)

/**
* RunQuery: Crea la consulta sobre model a partir de query y devuelve la página pedida.
* @param model *Model
* @param query et.Json   (where, order_by, page, limit)
* @return et.Items, error
**/
func RunQuery(model *Model, query et.Json) (et.Items, error) {
	ql := From(model)
	var err error

	ql, err = queryJoin(ql, query)
	if err != nil {
		return et.Items{}, err
	}

	ql, err = queryWhere(model, ql, query.ArrayJson("wheres"))
	if err != nil {
		return et.Items{}, err
	}

	ql, err = querySelect(ql, query.ArrayStr("selects"))
	if err != nil {
		return et.Items{}, err
	}

	ql, err = queryOrderBy(model, ql, query.Json("order_by"))
	if err != nil {
		return et.Items{}, err
	}

	page := query.Int("page")
	if page < 1 {
		page = 1
	}

	limit := query.Int("limit")
	if limit < 1 {
		limit = QUERY_DEFAULT_LIMIT
	}
	if limit > QUERY_MAX_LIMIT {
		limit = QUERY_MAX_LIMIT
	}

	return ql.Page(page).Rows(limit)
}

func whereElement(where et.Json) (field string, operator interface{}, value interface{}) {
	for column := range where {
		field = column
		w := where.Json(column)
		for op, val := range w {
			operator = op
			value = val
			break
		}
		break
	}

	return
}

/**
* queryJoin
* @param model *Model, ql *Ql, joins et.Json
* @return *Ql, error
**/
func queryJoin(ql *Ql, joins et.Json) (*Ql, error) {
	for _, kind := range []string{"join", "left_join", "right_join", "full_join"} {
		joins := joins.ArrayJson(kind)
		for _, join := range joins {
			for withName := range join {
				field, operator, value := whereElement(join.Json(withName))
				ql = ql.Join(withName, field, operator, value)
			}
		}
	}

	return ql, nil
}

/**
* queryWhere
* @param model *Model, ql *Ql, wheres []et.Json
* @return *Ql, error
**/
func queryWhere(model *Model, ql *Ql, wheres []et.Json) (*Ql, error) {
	applyWhere := func(where et.Json) {
		for column := range where {
			if strings.ToLower(column) == "and" {
				where = where.Json("and")
				for column = range where {
					field := model.GetField(column)
					if field == nil {
						continue
					}
					ql = ql.And(column)
					break
				}
			} else if strings.ToLower(column) == "or" {
				where = where.Json("or")
				for column = range where {
					field := model.GetField(column)
					if field == nil {
						continue
					}
					ql = ql.Or(column)
					break
				}
			} else {
				field := model.GetField(column)
				if field == nil {
					continue
				} 
				ql = ql.Where(column)
			}

			for operator, value := range where.Json(column) {
				switch operator {
				case "eq":
					ql = ql.Eq(value)
				case "neg":
					ql = ql.Neg(value)
				case "in":
					values, ok := value.([]interface{})
					if !ok {
						continue
					}
					ql = ql.In(values...)
				case "like":
					ql = ql.Like(value)
				case "more":
					ql = ql.More(value)
				case "less":
					ql = ql.Less(value)
				case "moreEq":
					ql = ql.MoreEq(value)
				case "lessEq":
					ql = ql.LessEq(value)
				case "isNull":
					ql = ql.IsNull()
				case "notNull":
					ql = ql.NotNull()
				default:
					continue
				}
			}
		}
	}

	for _, where := range wheres {
		applyWhere(where)
	}

	return ql, nil
}

/**
* querySelect
* @param ql *Ql, selects []string
* @return *Ql, error
**/
func querySelect(ql *Ql, selects []string) (*Ql, error) {
	for _, s := range selects {
		ql = ql.Select(s)
	}

	return ql, nil
}

/**
* queryOrderBy: Aplica el orden de order_by a ql, validando las columnas.
* @param model *Model, ql *Ql, orderBy et.Json
* @return *Ql, error
**/
func queryOrderBy(model *Model, ql *Ql, orderBy et.Json) (*Ql, error) {
	for _, direction := range []string{"asc", "desc"} {
		columns := orderBy.ArrayStr(direction)
		if len(columns) == 0 {
			continue
		}

		for _, column := range columns {
			if model.GetField(column) == nil {
				return nil, fmt.Errorf(MSG_QUERY_COLUMN_NOT_VALID, column)
			}
		}

		if direction == "asc" {
			ql = ql.OrderByAsc(columns...)
		} else {
			ql = ql.OrderByDesc(columns...)
		}
	}

	return ql, nil
}

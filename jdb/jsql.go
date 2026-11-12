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
* RunQuery: Crea la consulta sobre model a partir de query y devuelve la página pedida con el total de registros.
* @param model *Model
* @param query et.Json   (wheres, search, selects, order_by, page, rows; limit se acepta como alias de rows)
* @param rollups ...*Rollup   rollups extra (NewRollup) que se resuelven en cada fila
* @return et.List, error
**/
func RunQuery(model *Model, query et.Json, rollups ...*Rollup) (et.List, error) {
	ql := From(model).AddRollups(rollups...)
	var err error

	ql, err = queryJoin(ql, query)
	if err != nil {
		return et.List{}, err
	}

	ql, err = queryWhere(model, ql, query.ArrayJson("wheres"))
	if err != nil {
		return et.List{}, err
	}

	// search: texto libre sobre los campos de busqueda del modelo (DefineSearch), una condicion por palabra.
	search := query.Str("search")
	ql = ql.WhereSearch(search)

	ql, err = querySelect(ql, query.ArrayStr("selects"))
	if err != nil {
		return et.List{}, err
	}

	ql, err = queryOrderBy(model, ql, query.Json("order_by"))
	if err != nil {
		return et.List{}, err
	}

	page := query.Int("page")
	if page < 1 {
		page = 1
	}

	rows := query.Int("rows")
	if rows < 1 {
		rows = query.Int("limit")
	}
	if rows < 1 {
		rows = QUERY_DEFAULT_LIMIT
	}
	if rows > QUERY_MAX_LIMIT {
		rows = QUERY_MAX_LIMIT
	}

	return ql.List(page, rows)
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
	var applyOperators func(operators et.Json)

	// open: Abre la condicion de column con su conector (where, and, or) y aplica sus operadores. column es un campo
	// del modelo o concat ({ "fields": [...], "like": "%texto%" }: los campos unidos con un espacio). Lo que no se
	// reconoce se omite entero: sus operadores no deben caer sobre la condicion anterior.
	open := func(connector, column string, operators et.Json) bool {
		var fld interface{} = column
		if strings.EqualFold(column, "concat") {
			fields := []*Field{}
			for _, name := range operators.ArrayStr("fields") {
				if field := ql.getField(name); field != nil {
					fields = append(fields, field)
				}
			}
			if len(fields) == 0 {
				return false
			}
			fld = NewConcat(fields...)
			rest := et.Json{}
			for operator, value := range operators {
				if operator != "fields" {
					rest[operator] = value
				}
			}
			operators = rest
		} else if model.GetField(column) == nil {
			return false
		}

		if connector == "or" {
			ql = ql.Or(fld)
		} else {
			ql = ql.Where(fld)
		}
		applyOperators(operators)
		return true
	}

	applyWhere := func(where et.Json) {
		for column := range where {
			connector := strings.ToLower(column)
			if connector != "and" && connector != "or" {
				open("where", column, where.Json(column))
				continue
			}

			// { "and": { campo: {...} } } u { "or": ... }: la primera condicion reconocida
			group := where.Json(column)
			for inner := range group {
				if open(connector, inner, group.Json(inner)) {
					break
				}
			}
		}
	}

	applyOperators = func(operators et.Json) {
		for operator, value := range operators {
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
			case "search":
				ql = ql.Search(value)
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

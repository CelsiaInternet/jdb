package jdb

import (
	"fmt"
	"maps"
	"sort"

	"github.com/celsiainternet/elvis/console"
	"github.com/celsiainternet/elvis/et"
)

/**
* getDetailsTx: Completa la fila con lo que no sale del select: hoy los rollups pedidos (Select o Detail).
* @param tx *Tx, item et.Json
* @return et.Json, error
**/
func (s *Ql) getDetailsTx(tx *Tx, item et.Json) (et.Json, error) {
	for _, rollup := range s.Rollups {
		value, err := s.rollupValueTx(tx, rollup, item)
		if err != nil {
			return item, err
		}

		item[rollup.Name] = value
	}

	return item, nil
}

/**
* rollupValueTx: Valor de un rollup para la fila: consulta sus Fields en With por las llaves Fk.
* Con ShowAtrib y un solo campo devuelve el valor del campo; con varios campos, o con ShowObject, un objeto con los campos.
* Si a la fila le falta una llave o no hay registro relacionado, los campos quedan en nil.
* Las filas con las mismas llaves comparten la consulta (cache por Ql).
* @param tx *Tx, rollup *Rollup, item et.Json
* @return any, error
**/
func (s *Ql) rollupValueTx(tx *Tx, rollup *Rollup, item et.Json) (any, error) {
	with := rollup.getWith(s.Db)
	if with == nil {
		return nil, fmt.Errorf(MSG_MODEL_NOT_FOUND, rollup.WithName)
	}

	// Orden fijo de las llaves: arma el mismo where y la misma llave de cache para filas iguales.
	fkNames := make([]string, 0, len(rollup.Fk))
	for fkName := range rollup.Fk {
		fkNames = append(fkNames, fkName)
	}
	sort.Strings(fkNames)

	found := et.Json{}
	complete := true
	query := From(with)
	cacheKey := rollup.Name
	for idx, fkName := range fkNames {
		value, ok := item[fkName]
		if !ok || value == nil {
			complete = false
			break
		}

		pkName := rollup.Fk[fkName]
		cacheKey += fmt.Sprintf("|%s=%v", pkName, value)
		if idx == 0 {
			query.Where(pkName).Eq(value)
		} else {
			query.And(pkName).Eq(value)
		}
	}

	if complete {
		if cached, ok := s.rollupCache[cacheKey]; ok {
			return cloneRollupValue(cached), nil
		}

		// El select explicito no recorta las columnas en el driver: los campos se toman del resultado.
		row, err := query.OneTx(tx)
		if err != nil {
			return nil, err
		}

		if row.Ok {
			found = row.Result
		}
	}

	fields := et.Json{}
	for _, name := range rollup.Fields {
		fields[name] = found[name]
	}

	var result any = fields
	if rollup.Show == ShowAtrib && len(rollup.Fields) == 1 {
		result = fields[rollup.Fields[0]]
	}

	if complete {
		if s.rollupCache == nil {
			s.rollupCache = make(map[string]any)
		}
		s.rollupCache[cacheKey] = cloneRollupValue(result)
	}

	return result, nil
}

/**
* cloneRollupValue: Copia un objeto de rollup para que cada fila tenga el suyo (un valor simple se devuelve igual).
* @param value any
* @return any
**/
func cloneRollupValue(value any) any {
	if fields, ok := value.(et.Json); ok {
		return maps.Clone(fields)
	}

	return value
}

/**
* AllTx
* @param tx *Tx
* @return et.Items, error
**/
func (s *Ql) AllTx(tx *Tx) (et.Items, error) {
	if s.Db == nil {
		return et.Items{}, fmt.Errorf(MSG_DATABASE_IS_REQUIRED)
	}

	var err error
	var result et.Items
	if tx == nil {
		tx, err = s.beginTx()
		if err != nil {
			return et.Items{}, err
		}

		defer func() (et.Items, error) {
			if err == nil {
				err = tx.commit()
				if err != nil {
					return et.Items{}, err
				}
			}

			return result, err
		}()
	} else if s.tx == nil && tx != nil {
		s.tx = tx
	}

	result, err = s.Db.Select(s)
	if err != nil {
		return et.Items{}, err
	}

	for _, data := range result.Result {
		_, err = s.getDetailsTx(tx, data)
		if err != nil {
			return et.Items{}, err
		}
	}

	return result, nil
}

/**
* FirstTx
* @param tx *Tx, n int
* @return et.Item, error
**/
func (s *Ql) FirstTx(tx *Tx, n int) (et.Item, error) {
	result, err := s.AllTx(tx)
	if err != nil {
		return et.Item{}, err
	}

	if !result.Ok {
		return et.Item{}, nil
	}

	if n > result.Count {
		return et.Item{}, nil
	}

	if n < 0 {
		n = result.Count + n
	}

	return et.Item{
		Ok:     true,
		Result: result.Result[n],
	}, nil
}

/**
* LastTx
* @param tx *Tx, n int
* @return et.Items, error
**/
func (s *Ql) LastTx(tx *Tx, n int) (et.Item, error) {
	return s.FirstTx(tx, n*-1)
}

/**
* OneTx
* @param tx *Tx
* @return et.Item, error
**/
func (s *Ql) OneTx(tx *Tx) (et.Item, error) {
	result, err := s.FirstTx(tx, 0)
	if err != nil {
		return et.Item{}, err
	}

	return result, nil
}

/**
* RowsTx
* @param tx *Tx, limit int
* @return et.Items, error
**/
func (s *Ql) RowsTx(tx *Tx, val int) (et.Items, error) {
	s.Limit = val
	s.Offset = (s.Sheet - 1) * s.Limit
	if s.Offset < 0 {
		s.Offset = 0
	}
	return s.AllTx(tx)
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

	var err error
	var result bool
	if tx == nil {
		tx, err = s.beginTx()
		if err != nil {
			return false, err
		}

		defer func() (bool, error) {
			if err == nil {
				err = tx.commit()
				if err != nil {
					return false, err
				}
			}

			return result, err
		}()
	} else if s.tx == nil && tx != nil {
		s.tx = tx
	}

	result, err = s.Db.Exists(s)
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

	var err error
	var result int
	if tx == nil {
		tx, err = s.beginTx()
		if err != nil {
			return 0, err
		}

		defer func() (int, error) {
			if err == nil {
				err = tx.commit()
				if err != nil {
					return 0, err
				}
			}

			return result, err
		}()
	} else if s.tx == nil && tx != nil {
		s.tx = tx
	}

	result, err = s.Db.Count(s)
	if err != nil {
		return 0, err
	}

	return result, nil
}

/**
* All
* @return et.Items, error
**/
func (s *Ql) All() (et.Items, error) {
	return s.AllTx(nil)
}

/**
* First
* @param n int
* @return et.Item, error
**/
func (s *Ql) First(n int) (et.Item, error) {
	return s.FirstTx(nil, n)
}

/**
* Last
* @param n int
* @return et.Items, error
**/
func (s *Ql) Last(n int) (et.Item, error) {
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
		with := join.Str("with")
		field := join.Str("field")
		operator := StrToOperator(join.Str("operator"))
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

package oracle

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/celsiainternet/elvis/console"
	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/strs"
	"github.com/celsiainternet/jdb/jdb"
)

/**
* Exists
* @param ql *jdb.Ql
* @return bool, error
**/
func (s *Params) Exists(ql *jdb.Ql) (bool, error) {
	ql.Sql = ""
	ql.Sql = strs.Append(ql.Sql, "SELECT 1", "\n")
	ql.Sql = strs.Append(ql.Sql, s.sqlFrom(ql.Froms), "\n")
	ql.Sql = strs.Append(ql.Sql, s.sqlJoin(ql.Joins), "\n")
	ql.Sql = strs.Append(ql.Sql, s.sqlWhere(ql.QlWhere), "\n")
	ql.Sql = strs.Format(`SELECT CASE WHEN EXISTS (%s) THEN 1 ELSE 0 END AS "exists" FROM DUAL`, ql.Sql)

	if ql.IsDebug {
		console.Debug(ql.Sql)
	}

	items, err := jdb.QueryTx(s.jdb, ql.Tx(), ql.Sql)
	if err != nil {
		return false, err
	}

	if items.Count == 0 {
		return false, nil
	}

	return toInt(items.Result[0]["exists"]) == 1, nil
}

/**
* Count
* @param ql *jdb.Ql
* @return int, error
**/
func (s *Params) Count(ql *jdb.Ql) (int, error) {
	ql.Sql = ""
	ql.Sql = strs.Append(ql.Sql, `SELECT COUNT(*) AS "count"`, "\n")
	ql.Sql = strs.Append(ql.Sql, s.sqlFrom(ql.Froms), "\n")
	ql.Sql = strs.Append(ql.Sql, s.sqlJoin(ql.Joins), "\n")
	ql.Sql = strs.Append(ql.Sql, s.sqlWhere(ql.QlWhere), "\n")

	if ql.IsDebug {
		console.Debug(ql.Sql)
	}

	items, err := jdb.QueryTx(s.jdb, ql.Tx(), ql.Sql)
	if err != nil {
		return 0, err
	}

	if items.Count == 0 {
		return 0, nil
	}

	return toInt(items.Result[0]["count"]), nil
}

/**
* toInt
* @param val interface{}
* @return int
**/
func toInt(val interface{}) int {
	switch v := val.(type) {
	case int64:
		return int(v)
	case int:
		return v
	case float64:
		return int(v)
	case bool:
		if v {
			return 1
		}
		return 0
	default:
		var result int
		fmt.Sscan(fmt.Sprint(v), &result)
		return result
	}
}

/**
* sqlSelect
* @param ql *jdb.Ql
* @return string
**/
func (s *Params) sqlSelect(ql *jdb.Ql) string {
	if len(ql.Froms.Froms) == 0 {
		return ""
	}

	selects := []interface{}{}
	if len(ql.Selects) == 0 {
		from := ql.Froms.Froms[0]
		fields := from.GetColumnsFields()
		for _, field := range fields {
			selects = append(selects, field)
		}
	} else {
		selects = ql.Selects
	}

	var result string
	if ql.TypeSelect == jdb.Select {
		result = s.sqlColumns(selects)
	} else {
		result = s.sqlAtributes(selects)
	}

	result = strs.Append("\nSELECT", result, "\n")

	return result
}

/**
* asField: los atributos se leen con JSON_VALUE tipado (JSON_QUERY para objetos); los booleanos
* quedan como 1/0, igual que las columnas NUMBER(1).
* @param field jdb.Field
* @return string
**/
func asField(field jdb.Field) string {
	tableAlias := ""
	if field.Model != nil {
		tableAlias = field.Model.As
	}

	if len(field.Path) > 0 {
		// Ruta JSON dentro de la columna: JSON_VALUE la lee como texto.
		column := strs.Append(tableAlias, quoteIdent(field.Name), ".")
		return strs.Format(`JSON_VALUE(%s, %s)`, column, jsonPathKeys(field.Path))
	}

	if field.TypeColumn == jdb.TpColumn {
		return strs.Append(tableAlias, quoteIdent(field.Name), ".")
	} else if field.TypeColumn == jdb.TpAtribute && field.Model != nil && field.Model.SourceField != nil {
		source := strs.Append(tableAlias, quoteIdent(field.Model.SourceField.Name), ".")
		path := jsonPath(field.Name)
		def := quote(field.Default)
		switch field.TypeData {
		case jdb.TypeDataNumber, jdb.TypeDataInt, jdb.TypeDataPrecision:
			return strs.Format(`NVL(JSON_VALUE(%s, %s RETURNING NUMBER), %v)`, source, path, def)
		case jdb.TypeDataCheckbox:
			return strs.Format(`CASE JSON_VALUE(%s, %s) WHEN 'true' THEN 1 WHEN 'false' THEN 0 ELSE %v END`, source, path, def)
		case jdb.TypeDataObject, jdb.TypeDataMultiSelect, jdb.TypeDataGeometry:
			return strs.Format(`NVL(JSON_QUERY(%s, %s RETURNING CLOB), %v)`, source, path, def)
		default:
			return strs.Format(`NVL(JSON_VALUE(%s, %s), %v)`, source, path, def)
		}
	}

	return ""
}

/**
* aliasAsField
* @param field jdb.Field
* @return string
**/
func aliasAsField(field jdb.Field) string {
	result := asField(field)
	alias := field.As
	if alias == "" {
		alias = field.Name
	}
	if field.Name == alias {
		return result
	}

	return strs.Append(result, quoteIdent(alias), " AS ")
}

/**
* jsonExpr: adapta una expresión SQL para meterla en JSON_OBJECT conservando el tipo JSON:
* los booleanos (NUMBER(1)) salen como true/false y las columnas JSON como objeto, no como texto.
* @param expr string, typeData jdb.TypeData
* @return string
**/
func jsonExpr(expr string, typeData jdb.TypeData) string {
	switch typeData {
	case jdb.TypeDataCheckbox:
		return strs.Format(`CASE WHEN %s IS NULL THEN NULL WHEN %s = 1 THEN 'true' ELSE 'false' END FORMAT JSON`, expr, expr)
	case jdb.TypeDataObject, jdb.TypeDataMultiSelect, jdb.TypeDataGeometry:
		return strs.Format(`%s FORMAT JSON`, expr)
	default:
		return expr
	}
}

/**
* jsonBuild: objeto JSON con los pares; si hay source se combina con JSON_MERGEPATCH (19c no tiene
* JSON_TRANSFORM). El source va como parche para que una clave nula de las columnas no se pierda.
* @param source string, pairs []string
* @return string
**/
func jsonBuild(source string, pairs []string) string {
	object := `'{}'`
	if len(pairs) > 0 {
		object = strs.Format("JSON_OBJECT(\n%s\nNULL ON NULL RETURNING CLOB)", strings.Join(pairs, ",\n"))
	}

	if source == "" {
		return object
	}

	return strs.Format(`JSON_MERGEPATCH(%s, NVL(%s, '{}') RETURNING CLOB)`, object, source)
}

/**
* jsonPair: par 'key' VALUE expr para JSON_OBJECT.
* @param key string, expr string
* @return string
**/
func jsonPair(key, expr string) string {
	return strs.Format(`%v VALUE %s`, quote(key), expr)
}

/**
* agregationValue
* @param val interface{}
* @return string
**/
func agregationValue(val interface{}) string {
	switch v := val.(type) {
	case *jdb.Field:
		return asField(*v)
	case jdb.Field:
		return asField(v)
	case string:
		return v
	default:
		return fmt.Sprintf(`%v`, quote(v))
	}
}

/**
* agregationName
* @param agg *jdb.Agregation
* @return string
**/
func agregationName(agg *jdb.Agregation) string {
	switch agg.Agregation {
	case jdb.AgregationSum:
		return "SUM"
	case jdb.AgregationCount:
		return "COUNT"
	case jdb.AgregationAvg:
		return "AVG"
	case jdb.AgregationMin:
		return "MIN"
	case jdb.AgregationMax:
		return "MAX"
	case jdb.CommandExtractYear:
		return "YEAR"
	case jdb.CommandExtractMonth:
		return "MONTH"
	case jdb.CommandExtractDay:
		return "DAY"
	case jdb.CommandExtractHour:
		return "HOUR"
	case jdb.CommandExtractMinute:
		return "MINUTE"
	case jdb.CommandExtractSecond:
		return "SECOND"
	default:
		return ""
	}
}

/**
* asAgregation
* @param agg *jdb.Agregation
* @return string
**/
func asAgregation(agg *jdb.Agregation) string {
	val := agregationValue(agg.Value)
	switch agg.Agregation {
	case jdb.AgregationSum:
		return strs.Format(`SUM(%s)`, val)
	case jdb.AgregationCount:
		return strs.Format(`COUNT(%s)`, val)
	case jdb.AgregationAvg:
		return strs.Format(`AVG(%s)`, val)
	case jdb.AgregationMin:
		return strs.Format(`MIN(%s)`, val)
	case jdb.AgregationMax:
		return strs.Format(`MAX(%s)`, val)
	case jdb.CommandExtractYear:
		return strs.Format(`EXTRACT(YEAR FROM %s)`, val)
	case jdb.CommandExtractMonth:
		return strs.Format(`EXTRACT(MONTH FROM %s)`, val)
	case jdb.CommandExtractDay:
		return strs.Format(`EXTRACT(DAY FROM %s)`, val)
	case jdb.CommandExtractHour:
		return strs.Format(`EXTRACT(HOUR FROM %s)`, val)
	case jdb.CommandExtractMinute:
		return strs.Format(`EXTRACT(MINUTE FROM %s)`, val)
	case jdb.CommandExtractSecond:
		return strs.Format(`EXTRACT(SECOND FROM %s)`, val)
	default:
		return val
	}
}

/**
* aliasAsAgregation
* @param agg *jdb.Agregation
* @return string
**/
func aliasAsAgregation(agg *jdb.Agregation) string {
	result := asAgregation(agg)
	if agg.As == "" {
		return result
	}

	return strs.Append(result, quoteIdent(agg.As), " AS ")
}

/**
* sqlColumns
* @param selects []interface{}
* @return string
**/
func (s *Params) sqlColumns(selects []interface{}) string {
	if len(selects) == 0 {
		return "*"
	}

	result := ""
	for _, sel := range selects {
		switch v := sel.(type) {
		case *jdb.Field:
			result = strs.Append(result, aliasAsField(*v), ",\n")
		case jdb.Field:
			result = strs.Append(result, aliasAsField(v), ",\n")
		case *jdb.Agregation:
			result = strs.Append(result, aliasAsAgregation(v), ",\n")
		case jdb.Agregation:
			result = strs.Append(result, aliasAsAgregation(&v), ",\n")
		}
	}

	return result
}

/**
* sqlAtributes: proyección "source" (un solo objeto JSON por fila, en la columna result).
* @param selects []interface{}
* @return string
**/
func (s *Params) sqlAtributes(selects []interface{}) string {
	var sourceField *jdb.Field
	fields := []*jdb.Field{}
	aggs := []*jdb.Agregation{}
	for _, sel := range selects {
		switch v := sel.(type) {
		case *jdb.Field:
			fields = append(fields, v)
		case jdb.Field:
			fields = append(fields, &v)
		case *jdb.Agregation:
			aggs = append(aggs, v)
		case jdb.Agregation:
			aggs = append(aggs, &v)
		}
	}

	for _, field := range fields {
		if field.Name == jdb.SOURCE {
			sourceField = field
		}
	}

	source := ""
	if sourceField != nil {
		source = asField(*sourceField)
	}

	pairs := []string{}
	for _, field := range fields {
		if field == sourceField {
			continue
		}

		key := field.Name
		if field.As != "" {
			key = field.As
		}
		pairs = append(pairs, jsonPair(key, jsonExpr(asField(*field), field.TypeData)))
	}
	for _, agg := range aggs {
		key := agg.As
		if key == "" {
			key = agregationName(agg)
		}
		pairs = append(pairs, jsonPair(key, asAgregation(agg)))
	}

	return strs.Format(`%s AS "result"`, jsonBuild(source, pairs))
}

/**
* sqlFrom
* @param froms *jdb.QlFroms
* @return string
**/
func (s *Params) sqlFrom(froms *jdb.QlFroms) string {
	if len(froms.Froms) == 0 {
		return ""
	}

	return strs.Format("FROM %s", s.tableAs(froms.Froms[0]))
}

/**
* tableAs
* @param from *jdb.QlFrom
* @return string
**/
func (s *Params) tableAs(from *jdb.QlFrom) string {
	if from == nil {
		return ""
	}

	// Oracle no admite AS en el alias de tabla
	return strs.Append(tableName(from.Model), from.As, " ")
}

/**
* sqlJoin
* @param joins []*jdb.QlJoin
* @return string
**/
func (s *Params) sqlJoin(joins []*jdb.QlJoin) string {
	result := ""
	for _, join := range joins {
		if len(join.Condition) == 0 {
			continue
		}

		on := whereConditions(&jdb.QlWhere{Wheres: join.Condition})
		def := strs.Format("%s ON %s", s.tableAs(join.With), on)
		switch join.TypeJoin {
		case jdb.InnerJoin:
			result = strs.Append(result, strs.Append(`INNER JOIN`, def, " "), "\n")
		case jdb.LeftJoin:
			result = strs.Append(result, strs.Append(`LEFT JOIN`, def, " "), "\n")
		case jdb.RightJoin:
			result = strs.Append(result, strs.Append(`RIGHT JOIN`, def, " "), "\n")
		case jdb.FullJoin:
			result = strs.Append(result, strs.Append(`FULL JOIN`, def, " "), "\n")
		}
	}

	return result
}

/**
* sqlWhere
* @param where *jdb.QlWhere
* @return string
**/
func (s *Params) sqlWhere(where *jdb.QlWhere) string {
	if where == nil || len(where.Wheres) == 0 {
		return ""
	}

	return strs.Append("WHERE", whereConditions(where), " ")
}

/**
* whereConditions
* @param where *jdb.QlWhere
* @return string
**/
func whereConditions(where *jdb.QlWhere) string {
	result := ""
	for _, con := range where.Wheres {
		def := whereCondition(con)
		result = strs.Append(result, def, whereConnector(con.Connector))
	}

	return result
}

/**
* whereCondition
* @param con *jdb.QlCondition
* @return string
**/
func whereCondition(con *jdb.QlCondition) string {
	if con == nil {
		return ""
	}

	key := whereValue(con.Field)
	if con.Operator == jdb.Between {
		return strs.Format("%v%v", key, whereBetween(con.Value))
	}

	values := whereValue(con.Value)
	if con.Operator == jdb.Like || con.Operator == jdb.Search {
		key = strs.Format(`UPPER(%v)`, key)
	}

	return strs.Format("%v%v", key, whereOperator(con, values))
}

/**
* whereBetween
* @param val interface{}
* @return string
**/
func whereBetween(val interface{}) string {
	items := arrayItems(val)
	if len(items) != 2 {
		return ""
	}

	return strs.Format(" BETWEEN %v AND %v", whereValue(items[0]), whereValue(items[1]))
}

/**
* arrayItems
* @param val interface{}
* @return []interface{}
**/
func arrayItems(val interface{}) []interface{} {
	switch v := val.(type) {
	case jdb.Value:
		return arrayItems(v.Value)
	case *jdb.Value:
		return arrayItems(v.Value)
	}

	rv := reflect.ValueOf(val)
	if rv.Kind() != reflect.Slice {
		return nil
	}

	result := make([]interface{}, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		result[i] = rv.Index(i).Interface()
	}

	return result
}

/**
* whereValueTyped
* @param v jdb.Value
* @return string
**/
func whereValueTyped(v jdb.Value) string {
	if v.Type == jdb.ValueTypeArray {
		if items := arrayItems(v.Value); items != nil {
			parts := make([]string, len(items))
			for i, item := range items {
				parts[i] = whereValue(item)
			}
			return strings.Join(parts, ", ")
		}
	}

	return strs.Format(`%v`, quote(v))
}

/**
* whereValue
* @param val interface{}
* @return string
**/
func whereValue(val interface{}) string {
	switch v := val.(type) {
	case jdb.Value:
		return whereValueTyped(v)
	case *jdb.Value:
		return whereValueTyped(*v)
	case jdb.Field:
		return asField(v)
	case *jdb.Field:
		return asField(*v)
	case jdb.Column:
		return quoteIdent(v.Name)
	case *jdb.Column:
		return quoteIdent(v.Name)
	case []interface{}, []string, []int, []float64:
		items := arrayItems(v)
		parts := make([]string, len(items))
		for i, item := range items {
			parts[i] = whereValue(item)
		}
		return strings.Join(parts, ", ")
	default:
		return strs.Format(`%v`, quote(v))
	}
}

/**
* whereOperator: LIKE se compara en mayúsculas (whereCondition envuelve el campo en UPPER); Search
* no usa Oracle Text y se resuelve como LIKE '%valor%'.
* @param condition *jdb.QlCondition, val interface{}
* @return string
**/
func whereOperator(condition *jdb.QlCondition, val interface{}) string {
	switch condition.Operator {
	case jdb.Equal:
		return strs.Format("=%v", val)
	case jdb.Neg:
		return strs.Format("!=%v", val)
	case jdb.In:
		return strs.Format(" IN (%v)", val)
	case jdb.NotIn:
		return strs.Format(" NOT IN (%v)", val)
	case jdb.Like:
		return strs.Format(" LIKE UPPER(%v)", val)
	case jdb.More:
		return strs.Format(">%v", val)
	case jdb.Less:
		return strs.Format("<%v", val)
	case jdb.MoreEq:
		return strs.Format(">=%v", val)
	case jdb.LessEq:
		return strs.Format("<=%v", val)
	case jdb.IsNull:
		return " IS NULL"
	case jdb.NotNull:
		return " IS NOT NULL"
	case jdb.Search:
		return strs.Format(" LIKE '%%' || UPPER(%v) || '%%'", val)
	default:
		return ""
	}
}

func whereConnector(con jdb.Connector) string {
	switch con {
	case jdb.And:
		return "\nAND "
	case jdb.Or:
		return "\nOR "
	default:
		return ""
	}
}

/**
* sqlGroupBy: sin alias, GROUP BY no admite "expr AS alias".
* @param ql *jdb.Ql
* @return string
**/
func (s *Params) sqlGroupBy(ql *jdb.Ql) string {
	result := ""
	for _, field := range ql.Groups {
		result = strs.Append(result, asField(*field), ",\n")
	}

	if result == "" {
		return ""
	}

	return strs.Format("GROUP BY %s", result)
}

/**
* sqlHaving
* @param ql *jdb.Ql
* @return string
**/
func (s *Params) sqlHaving(ql *jdb.Ql) string {
	if ql.Havings == nil {
		return ""
	}

	where := whereConditions(ql.Havings)
	if where == "" {
		return ""
	}

	return strs.Format("HAVING %s", where)
}

/**
* sqlOrderBy
* @param ql *jdb.Ql
* @return string
**/
func (s *Params) sqlOrderBy(ql *jdb.Ql) string {
	if ql.Orders == nil {
		return ""
	}

	result := ""
	for _, fld := range ql.Orders.Asc {
		result = strs.Append(result, strs.Append(asField(*fld), "ASC", " "), ",\n")
	}
	for _, fld := range ql.Orders.Desc {
		result = strs.Append(result, strs.Append(asField(*fld), "DESC", " "), ",\n")
	}

	if len(result) != 0 {
		result = strs.Append("ORDER BY", result, "\n")
	}

	return result
}

/**
* sqlLimit: OFFSET/FETCH de Oracle 12c+.
* @param ql *jdb.Ql
* @return string
**/
func (s *Params) sqlLimit(ql *jdb.Ql) string {
	if ql.Sheet > 0 {
		return strs.Format(`OFFSET %d ROWS FETCH NEXT %d ROWS ONLY`, ql.Offset, ql.Limit)
	} else if ql.Limit > 0 {
		return strs.Format(`FETCH FIRST %d ROWS ONLY`, ql.Limit)
	}

	return ""
}

/**
* Select
* @param ql *jdb.Ql
* @return et.Items, error
**/
func (s *Params) Select(ql *jdb.Ql) (et.Items, error) {
	ql.Sql = ""
	ql.Sql = strs.Append(ql.Sql, s.sqlSelect(ql), "\n")
	ql.Sql = strs.Append(ql.Sql, s.sqlFrom(ql.Froms), "\n")
	ql.Sql = strs.Append(ql.Sql, s.sqlJoin(ql.Joins), "\n")
	ql.Sql = strs.Append(ql.Sql, s.sqlWhere(ql.QlWhere), "\n")
	ql.Sql = strs.Append(ql.Sql, s.sqlGroupBy(ql), "\n")
	ql.Sql = strs.Append(ql.Sql, s.sqlHaving(ql), "\n")
	ql.Sql = strs.Append(ql.Sql, s.sqlOrderBy(ql), "\n")
	ql.Sql = strs.Append(ql.Sql, s.sqlLimit(ql), "\n")

	if ql.IsDebug {
		console.Debug(ql.Sql)
	}

	result, err := jdb.QueryTx(s.jdb, ql.Tx(), ql.Sql)
	if err != nil {
		return et.Items{}, err
	}

	if ql.TypeSelect == jdb.Select {
		return normalizeColumns(result, selectTypes(ql)), nil
	}

	return normalizeResult(result), nil
}

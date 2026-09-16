package postgres

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/celsiainternet/elvis/console"
	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/strs"
	jdb "github.com/celsiainternet/jdb/jdb"
)

/**
* Exists
* @param ql *jdb.Ql
* @return bool, error
**/
func (s *Postgres) Exists(ql *jdb.Ql) (bool, error) {
	ql.Sql = ""
	ql.Sql = strs.Append(ql.Sql, "SELECT 1", "\n")
	ql.Sql = strs.Append(ql.Sql, s.sqlFrom(ql.Froms), "\n")
	ql.Sql = strs.Append(ql.Sql, s.sqlJoin(ql.Joins), "\n")
	ql.Sql = strs.Append(ql.Sql, s.sqlWhere(ql.QlWhere), "\n")

	if len(ql.Sql) > 0 {
		ql.Sql = strs.Format("SELECT EXISTS (%s);", ql.Sql)
	}

	if ql.IsDebug {
		console.Debug(ql.Sql)
	}

	item, err := jdb.Query(s.jdb, ql.Sql)
	if err != nil {
		return false, err
	}

	result := item.Bool(0, "exists")

	return result, nil
}

/**
* Count
* @param ql *jdb.Ql
* @return int, error
**/
func (s *Postgres) Count(ql *jdb.Ql) (int, error) {
	ql.Sql = ""
	ql.Sql = strs.Append(ql.Sql, "SELECT COUNT(*) AS Count", "\n")
	ql.Sql = strs.Append(ql.Sql, s.sqlFrom(ql.Froms), "\n")
	ql.Sql = strs.Append(ql.Sql, s.sqlJoin(ql.Joins), "\n")
	ql.Sql = strs.Append(ql.Sql, s.sqlWhere(ql.QlWhere), "\n")

	if ql.IsDebug {
		console.Debug(ql.Sql)
	}

	result, err := jdb.Query(s.jdb, ql.Sql)
	if err != nil {
		return 0, err
	}

	if result.Count == 0 {
		return 0, nil
	}

	return result.Int(0, "count"), nil
}

/**
* sqlSelect
* @param ql *jdb.Ql
* @return string
**/
func (s *Postgres) sqlSelect(ql *jdb.Ql) string {
	if len(ql.Froms.Froms) == 0 {
		return ""
	}

	var result string
	if ql.TypeSelect == jdb.Select {
		result = s.sqlColumns(ql.Selects)
	} else {
		result = s.sqlAtributes(ql.Selects)
	}

	result = strs.Append("\nSELECT", result, "\n")

	return result
}

/**
* asField renders the raw SQL reference used to read a field's value:
* "alias.column" for a real column, or the jsonb attribute-extraction
* expression (with COALESCE to the column's default) for a TpAtribute field
* backed by the model's SourceField. Aggregation wrapping (SUM/COUNT/EXTRACT/
* ...) is handled separately by asAgregation, since jdb.Field carries no
* aggregation information of its own.
* @param field jdb.Field
* @return string
**/
func asField(field jdb.Field) string {
	tableAlias := ""
	if field.Model != nil {
		tableAlias = field.Model.As
	}

	if field.TypeColumn == jdb.TpAtribute && field.Model != nil && field.Model.SourceField != nil {
		source := strs.Append(tableAlias, field.Model.SourceField.Name, ".")
		result := strs.Format(`%s#>>'{%s}'`, source, field.Name)
		result = strs.Format(`COALESCE(%s, %v)`, result, quote(field.Default))
		return result
	}

	return strs.Append(tableAlias, field.Name, ".")
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

	return strs.Append(result, alias, " AS ")
}

/**
* agregationValue renders the operand of an aggregation/calc expression: a
* *jdb.Field/jdb.Field is rendered like any other selected column, a plain
* string is treated as a raw SQL expression or column reference (this is how
* CALC's expression is meant to be embedded), and anything else is rendered
* as a SQL literal.
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
* asAgregation renders a *jdb.Agregation (SUM, COUNT, AVG, MIN, MAX, the
* CommandExtractYear/Month/Day/Hour/Minute/Second family, VALUE and CALC)
* as a raw SQL expression.
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

	return strs.Append(result, agg.As, " AS ")
}

/**
* sqlBuildObject
* @param selects []*jdb.Field
* @return string
**/
func (s *Postgres) sqlBuildObject(selects []*jdb.Field) string {
	result := ""
	l := 20
	if s.version >= 13 {
		l = 100
	}
	n := 0
	obj := ""
	sourceField := make([]*jdb.Field, 0)
	for _, fld := range selects {
		n++
		if fld.Model != nil && fld.Model.SourceField != nil && fld.Name == fld.Model.SourceField.Name {
			sourceField = append(sourceField, fld)
			continue
		}
		def := asField(*fld)
		alias := fld.As
		if def == "" || alias == "" {
			continue
		}
		def = strs.Format(`'%s', %s`, alias, def)
		obj = strs.Append(obj, def, ",\n")

		if n == l {
			result = jsonBuildObject(result, obj)
			obj = ""
			n = 0
		}
	}
	if n > 0 {
		result = jsonBuildObject(result, obj)
	}
	sources := ""
	for i := 0; i < len(sourceField); i++ {
		fld := sourceField[i]
		def := aliasAsField(*fld)
		sources = strs.Append(sources, def, "||\n")
		if i == len(sourceField)-1 {
			result = strs.Format(`%s||%s`, def, result)
		}
	}

	return result
}

/**
* jsonBuildObject
* @param result, obj string
* @return string
**/
func jsonBuildObject(result, obj string) string {
	if len(obj) == 0 {
		return result
	}

	return strs.Append(result, strs.Format("jsonb_build_object(\n%s)", obj), "||\n")
}

/**
* sqlAtributes renders a Ql.Selects-style list (a mix of *jdb.Field and
* *jdb.Agregation entries, as produced by Ql.Select/Ql.Data) as the "source"
* (jsonb) projection: the plain fields are merged into the source jsonb blob
* via sqlBuildObject, and any explicitly selected aggregation is appended as
* an extra key in the resulting object.
* @param selects []interface{}
* @return string
**/
func (s *Postgres) sqlAtributes(selects []interface{}) string {
	fields := []*jdb.Field{}
	extra := []string{}
	for _, sel := range selects {
		switch v := sel.(type) {
		case *jdb.Field:
			fields = append(fields, v)
		case jdb.Field:
			fields = append(fields, &v)
		case *jdb.Agregation:
			extra = append(extra, strs.Format(`'%s', %s`, v.As, asAgregation(v)))
		case jdb.Agregation:
			extra = append(extra, strs.Format(`'%s', %s`, v.As, asAgregation(&v)))
		}
	}

	result := s.sqlBuildObject(fields)
	if len(extra) > 0 {
		extraObj := strs.Format("jsonb_build_object(\n%s)", strings.Join(extra, ",\n"))
		if result == "" {
			result = extraObj
		} else {
			result = strs.Format("%s||%s", result, extraObj)
		}
	}
	result = strs.Append(result, "result", " AS ")

	return result
}

/**
* sqlColumns renders a Ql.Selects-style list (a mix of *jdb.Field and
* *jdb.Agregation entries) as a comma-separated column list. Also used for
* plain *jdb.Field lists such as Ql.Groups.
* @param selects []interface{}
* @return string
**/
func (s *Postgres) sqlColumns(selects []interface{}) string {
	result := ""
	for _, sel := range selects {
		switch v := sel.(type) {
		case *jdb.Field:
			def := aliasAsField(*v)
			result = strs.Append(result, def, ",\n")
		case jdb.Field:
			def := aliasAsField(v)
			result = strs.Append(result, def, ",\n")
		case *jdb.Agregation:
			def := aliasAsAgregation(v)
			result = strs.Append(result, def, ",\n")
		case jdb.Agregation:
			def := aliasAsAgregation(&v)
			result = strs.Append(result, def, ",\n")
		}
	}

	return result
}

/**
* fieldsToSelects converts a plain []*jdb.Field (e.g. Ql.Groups) into the
* []interface{} shape shared with Ql.Selects, so both can be rendered by the
* same sqlColumns.
* @param fields []*jdb.Field
* @return []interface{}
**/
func fieldsToSelects(fields []*jdb.Field) []interface{} {
	result := make([]interface{}, len(fields))
	for i, field := range fields {
		result[i] = field
	}

	return result
}

/**
* sqlFrom
* @param froms *jdb.QlFroms
* @return string
**/
func (s *Postgres) sqlFrom(froms *jdb.QlFroms) string {
	if len(froms.Froms) == 0 {
		return ""
	}

	from := froms.Froms[0]
	def := s.tableAs(from)
	result := strs.Format("FROM %s", def)

	return result
}

/**
* tableAs
* @param from *jdb.QlFrom
* @return string
**/
func (s *Postgres) tableAs(from *jdb.QlFrom) string {
	if from == nil {
		return ""
	}

	table := tableName(from.Model)
	return strs.Append(table, from.As, " AS ")
}

/**
* sqlJoin renders each join's ON clause by reusing whereConditions over its
* jdb.QlCondition slice - a join's condition is a QlCondition just like a
* WHERE clause's, so the same field/operator/value rendering applies.
* @param joins []*jdb.QlJoin
* @return string
**/
func (s *Postgres) sqlJoin(joins []*jdb.QlJoin) string {
	result := ""
	for _, join := range joins {
		if len(join.Condition) == 0 {
			continue
		}

		on := whereConditions(&jdb.QlWhere{Wheres: join.Condition})
		def := strs.Format("%s ON %s", s.tableAs(join.With), on)
		switch join.TypeJoin {
		case jdb.InnerJoin:
			def = strs.Append(`INNER JOIN`, def, " ")
			result = strs.Append(result, def, "\n")
		case jdb.LeftJoin:
			def = strs.Append(`LEFT JOIN`, def, " ")
			result = strs.Append(result, def, "\n")
		case jdb.RightJoin:
			def = strs.Append(`RIGHT JOIN`, def, " ")
			result = strs.Append(result, def, "\n")
		case jdb.FullJoin:
			def = strs.Append(`FULL JOIN`, def, " ")
			result = strs.Append(result, def, "\n")
		}
	}

	return result
}

/**
* sqlWhere
* @param where *jdb.QlWhere
* @return string
**/
func (s *Postgres) sqlWhere(where *jdb.QlWhere) string {
	if where == nil {
		return ""
	}

	if len(where.Wheres) == 0 {
		return ""
	}

	result := whereConditions(where)
	result = strs.Append("WHERE", result, " ")

	return result
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
		conector := whereConnector(con.Connector)
		result = strs.Append(result, def, conector)
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
	def := whereOperator(con, values)
	return strs.Format("%v%v", key, def)
}

/**
* whereBetween renders a BETWEEN condition's two bounds as valid Postgres
* syntax ("BETWEEN low AND high"). QlWhere.Between wraps its bounds as a
* two-element ValueTypeArray, which the generic "operator (%v)" rendering
* whereOperator uses for every other operator can't turn into "AND"-joined
* bounds, so BETWEEN is rendered separately here.
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
* arrayItems unwraps a jdb.Value-wrapped array (or a bare slice, of any
* element type) into its elements. Returns nil if val is neither.
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
* whereValueTyped renders a jdb.Value according to its tagged ValueType:
* ValueTypeArray (the .In/.NotIn operand) expands to a comma-separated list
* of its elements, suitable for "IN (...)"/"NOT IN (...)". Every other
* ValueType (a single scalar/Json/Field/Agregation/Calc value) renders as one
* value via quote(), which already knows how to unwrap a jdb.Value the same
* way (see quoteValue in quote.go) - keeping ValueTypeJsonArray, for example,
* as a single JSON blob rather than expanding it like ValueTypeArray.
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
		return v.Name
	case *jdb.Column:
		return v.Name
	case []interface{}:
		parts := make([]string, len(v))
		for i, vl := range v {
			vs := whereValue(vl)
			parts[i] = fmt.Sprint(vs)
		}
		vals := strings.Join(parts, ", ")
		return vals
	case []string:
		parts := make([]string, len(v))
		for i, vl := range v {
			vs := whereValue(vl)
			parts[i] = fmt.Sprint(vs)
		}
		vals := strings.Join(parts, ", ")
		return vals
	case []int:
		parts := make([]string, len(v))
		for i, vl := range v {
			parts[i] = fmt.Sprint(vl)
		}
		vals := strings.Join(parts, ", ")
		return vals
	case []float64:
		parts := make([]string, len(v))
		for i, vl := range v {
			parts[i] = fmt.Sprint(vl)
		}
		vals := strings.Join(parts, ", ")
		return vals
	default:
		return strs.Format(`%v`, quote(v))
	}
}

/**
* whereOperator
* @param condition *jdb.QlCondition
* @param val interface{}
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
		return strs.Format(" ILIKE %v", val)
	case jdb.More:
		return strs.Format(">%v", val)
	case jdb.Less:
		return strs.Format("<%v", val)
	case jdb.MoreEq:
		return strs.Format(">=%v", val)
	case jdb.LessEq:
		return strs.Format("<=%v", val)
	case jdb.Between:
		// Unreachable: whereCondition intercepts jdb.Between and renders it
		// via whereBetween before ever calling whereOperator, since BETWEEN
		// needs its two bounds joined with "AND", not the generic "(%v)"
		// wrapping every other operator here uses.
		return ""
	case jdb.IsNull:
		return " IS NULL"
	case jdb.NotNull:
		return " IS NOT NULL"
	case jdb.Search:
		return strs.Format(" @@ to_tsquery('%s', %v)", searchLanguage(condition.Field), val)
	default:
		return ""
	}
}

/**
* searchLanguage resolves the text-search language ("english" by default,
* postgres' own default config) for a full-text Search condition. jdb's
* QlCondition carries no language of its own - the FullText language, if
* declared, lives on the model of the field being searched.
* @param field interface{}
* @return string
**/
func searchLanguage(field interface{}) string {
	var target *jdb.Field
	switch v := field.(type) {
	case *jdb.Field:
		target = v
	case jdb.Field:
		target = &v
	default:
		return "english"
	}

	if target == nil || target.Model == nil {
		return "english"
	}

	for _, ft := range target.Model.FullText {
		if ft.Language != "" {
			return ft.Language
		}
	}

	return "english"
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
* sqlGroupBy
* @param ql *jdb.Ql
* @return string
**/
func (s *Postgres) sqlGroupBy(ql *jdb.Ql) string {
	result := ""
	columns := s.sqlColumns(fieldsToSelects(ql.Groups))
	if len(columns) == 0 {
		return result
	}

	result = strs.Format("GROUP BY %s", columns)

	return result
}

/**
* sqlHaving
* @param ql *jdb.Ql
* @return string
**/
func (s *Postgres) sqlHaving(ql *jdb.Ql) string {
	result := ""
	havings := ql.Havings
	where := whereConditions(havings)
	if where == "" {
		return result
	}

	result = strs.Format("HAVING %s", where)

	return result
}

/**
* sqlOrderBy
* @param ql *jdb.Ql
* @return string
**/
func (s *Postgres) sqlOrderBy(ql *jdb.Ql) string {
	result := ""
	for _, fld := range ql.Orders.Asc {
		def := asField(*fld)
		def = strs.Append(def, "ASC", " ")
		result = strs.Append(result, def, ",\n")
	}
	for _, fld := range ql.Orders.Desc {
		def := asField(*fld)
		def = strs.Append(def, "DESC", " ")
		result = strs.Append(result, def, ",\n")
	}

	if len(result) != 0 {
		result = strs.Append("ORDER BY", result, "\n")
	}

	return result
}

/**
* sqlLimit
* @param ql *jdb.Ql
* @return string
**/
func (s *Postgres) sqlLimit(ql *jdb.Ql) string {
	result := ""
	if ql.Sheet > 0 {
		result = strs.Format(`LIMIT %d OFFSET %d`, ql.Limit, ql.Offset)
	} else if ql.Limit > 0 {
		result = strs.Format(`LIMIT %d`, ql.Limit)
	}

	return result
}

/**
* Select
* @param ql *jdb.Ql
* @return et.Items, error
**/
func (s *Postgres) Select(ql *jdb.Ql) (et.Items, error) {
	ql.Sql = ""
	ql.Sql = strs.Append(ql.Sql, s.sqlSelect(ql), "\n")
	ql.Sql = strs.Append(ql.Sql, s.sqlFrom(ql.Froms), "\n")
	ql.Sql = strs.Append(ql.Sql, s.sqlJoin(ql.Joins), "\n")
	ql.Sql = strs.Append(ql.Sql, s.sqlWhere(ql.QlWhere), "\n")
	ql.Sql = strs.Append(ql.Sql, s.sqlGroupBy(ql), "\n")
	ql.Sql = strs.Append(ql.Sql, s.sqlHaving(ql), "\n")
	ql.Sql = strs.Append(ql.Sql, s.sqlOrderBy(ql), "\n")
	ql.Sql = strs.Append(ql.Sql, s.sqlLimit(ql), "\n")
	ql.Sql = strs.Format(`%s;`, ql.Sql)

	if ql.IsDebug {
		console.Debug(ql.Sql)
	}

	result, err := jdb.QueryTx(s.jdb, ql.Tx(), ql.Sql)
	if err != nil {
		return et.Items{}, err
	}

	return result, nil
}

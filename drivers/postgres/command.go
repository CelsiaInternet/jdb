package postgres

import (
	"fmt"
	"slices"
	"strings"

	"github.com/celsiainternet/elvis/console"
	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/strs"
	jdb "github.com/celsiainternet/jdb/jdb"
)

/**
* jsonColumnTypes are the TypeData variants that ddl-table.go maps to a JSONB
* postgres column (see (*Postgres).typeData) and therefore need an explicit
* ::jsonb cast when their value is inlined as a text literal.
**/
var jsonColumnTypes = []jdb.TypeData{jdb.TypeDataObject, jdb.TypeDataMultiSelect, jdb.TypeDataGeometry}

/**
* returningColumns builds the "'name', name" pairs used inside
* jsonb_build_object(...) for a command's RETURNING clause: the caller's
* explicit Returns(...) projection when given (requested), otherwise every
* real column of the model - excluding the source/jsonb blob column, which
* every caller merges in separately via "||".
* @param from *jdb.QlFrom, requested []*jdb.Field
* @return []string
**/
func returningColumns(from *jdb.QlFrom, requested []*jdb.Field) []string {
	isReturnable := func(name string, typeColumn jdb.TypeColumn) bool {
		if from.SourceField != nil && name == from.SourceField.Name {
			return false
		}
		return typeColumn == jdb.TpColumn
	}

	result := []string{}
	if len(requested) > 0 {
		for _, field := range requested {
			if !isReturnable(field.Name, field.TypeColumn) {
				continue
			}
			result = append(result, strs.Format("'%s', %s", field.Name, field.Name))
		}
		return result
	}

	for _, column := range from.Model.Columns {
		if !isReturnable(column.Name, column.TypeColumn) {
			continue
		}
		result = append(result, strs.Format("'%s', %s", column.Name, column.Name))
	}

	return result
}

/**
* sqlInsert
* @param command *jdb.Command
* @return string, []any
**/
func (s *Postgres) sqlInsert(command *jdb.Command) (string, []any) {
	from := command.GetFrom()
	if from == nil {
		return "", []any{}
	}

	table := tableName(from.Model)
	columns := []string{}
	values := []string{}
	args := []any{}
	data := et.Json{}
	for key, val := range command.New {
		field := from.GetField(key)
		if field == nil {
			continue
		}

		switch field.TypeColumn {
		case jdb.TpColumn:
			if from.SourceField != nil && field.Name == from.SourceField.Name {
				continue
			}

			columns = append(columns, field.Name)
			val := quote(val)
			if slices.Contains(jsonColumnTypes, field.TypeData) {
				val = fmt.Sprintf(`%v::jsonb`, val)
			}
			values = append(values, strs.Format(`%v`, val))
		case jdb.TpAtribute:
			data.Set(key, val)
		}
	}

	returns := returningColumns(from, command.Returning)
	returnsStr := strings.Join(returns, ",")

	result := "INSERT INTO %s(\n%s)\nVALUES (%s)\nRETURNING\n%s AS result;"
	if from.SourceField != nil {
		columns = append(columns, from.SourceField.Name)
		arg := strs.Format(`%v`, data.ToString())
		args = append(args, arg)
		values = append(values, fmt.Sprintf(`$%d::jsonb`, len(args)))
		columnsStr := strings.Join(columns, ",\n")
		valuesStr := strings.Join(values, ",")
		returnsStr = fmt.Sprintf(`%s || jsonb_build_object(%s)`, from.SourceField.Name, returnsStr)
		result = strs.Format(result, table, columnsStr, valuesStr, returnsStr)
	} else {
		columnsStr := strings.Join(columns, ",\n")
		valuesStr := strings.Join(values, ",")
		returnsStr = fmt.Sprintf(`jsonb_build_object(%s)`, returnsStr)
		result = strs.Format(result, table, columnsStr, valuesStr, returnsStr)
	}

	return result, args
}

/**
* sqlUpdate
* @param command *jdb.Command
* @return string, []any
**/
func (s *Postgres) sqlUpdate(command *jdb.Command) (string, []any) {
	args := []any{}
	from := command.GetFrom()
	if from == nil {
		return "", args
	}

	set := []string{}
	_data := ""
	for key, val := range command.New {
		field := from.GetField(key)
		if field == nil {
			continue
		}

		switch field.TypeColumn {
		case jdb.TpColumn:
			if from.SourceField != nil && field.Name == from.SourceField.Name {
				continue
			}

			val := quote(val)
			if slices.Contains(jsonColumnTypes, field.TypeData) {
				val = fmt.Sprintf(`%v::jsonb`, val)
			}
			set = append(set, strs.Format(`%s = %v`, field.Name, val))
		case jdb.TpAtribute:
			val := quote(val)
			strVal := fmt.Sprintf(`%v`, val)
			if len(strVal) == 0 || strVal == "''" {
				continue
			}

			tp := s.typeData(field.TypeData)
			if len(_data) == 0 {
				_data = fmt.Sprintf("COALESCE(%s, '{}')", from.SourceField.Name)
				_data = strs.Format("jsonb_set(%s,\n'{%s}', to_jsonb(%v::%v), true)", _data, key, val, tp)
			} else {
				_data = strs.Format("jsonb_set(\n%s,\n'{%s}', to_jsonb(%v::%v), true)", _data, key, val, tp)
			}
		}
	}

	if len(set) == 0 && len(_data) == 0 {
		// Nothing to assign - "UPDATE table SET WHERE ..." is not valid SQL,
		// so signal "can't build this" the same way from == nil does above
		// rather than emit a broken statement.
		return "", args
	}

	returns := returningColumns(from, command.Returning)

	where := whereConditions(command.QlWhere)
	table := tableName(from.Model)
	result := "UPDATE %s\nSET\n%s\nWHERE %s\nRETURNING\njsonb_build_object(%s) AS result;"
	if from.SourceField != nil {
		if len(_data) > 0 {
			column := from.SourceField.Name
			set = append(set, strs.Format(`%s = %s`, column, _data))
		}
		result = "UPDATE %s\nSET\n%s\nWHERE %s\nRETURNING\n%s || jsonb_build_object(%s) AS result;"
		result = strs.Format(result, table, strings.Join(set, ",\n"), where, from.SourceField.Name, strings.Join(returns, ","))
	} else {
		result = strs.Format(result, table, strings.Join(set, ",\n"), where, strings.Join(returns, ","))
	}

	return result, args
}

/**
* sqlDelete
* @param command *jdb.Command
* @return string, []any
**/
func (s *Postgres) sqlDelete(command *jdb.Command) (string, []any) {
	args := []any{}
	from := command.GetFrom()
	if from == nil {
		return "", args
	}

	returns := returningColumns(from, command.Returning)

	where := whereConditions(command.QlWhere)
	table := tableName(from.Model)
	result := "DELETE FROM %s\nWHERE %s\nRETURNING\njsonb_build_object(%s) AS result;"
	if from.SourceField != nil {
		result = "DELETE FROM %s\nWHERE %s\nRETURNING\n%s || jsonb_build_object(%s) AS result;"
		result = strs.Format(result, table, where, from.SourceField.Name, strings.Join(returns, ","))
	} else {
		result = strs.Format(result, table, where, strings.Join(returns, ","))
	}
	return result, args
}

/**
* Command
* @param command *jdb.Command
* @return et.Items, error
**/
func (s *Postgres) Command(command *jdb.Command) (et.Items, error) {
	command.Sql = ""
	command.Args = []any{}
	switch command.Command {
	case jdb.Insert:
		sql, args := s.sqlInsert(command)
		command.Sql = strs.Append(command.Sql, sql, "\n")
		command.Args = append(command.Args, args...)
	case jdb.Update:
		sql, args := s.sqlUpdate(command)
		command.Sql = strs.Append(command.Sql, sql, "\n")
		command.Args = append(command.Args, args...)
	case jdb.Delete:
		sql, args := s.sqlDelete(command)
		command.Sql = strs.Append(command.Sql, sql, "\n")
		command.Args = append(command.Args, args...)
	}

	if command.IsDebug {
		console.Debug(et.Json{"sql": command.Sql, "args": command.Args}.ToString())
	}

	result, err := jdb.QueryTx(s.jdb, command.Tx(), command.Sql, command.Args...)
	if err != nil {
		console.Error(err)
		console.Debug(command.Sql)
		console.Debug(command.Args)
		return et.Items{}, err
	}

	return result, nil
}

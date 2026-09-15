package postgres

import (
	"fmt"
	"strings"

	"github.com/celsiainternet/elvis/console"
	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/strs"
	jdb "github.com/celsiainternet/jdb/jdb"
)

/**
* sqlInsert
* @param command *jdb.Command
* @return string
**/
func (s *Postgres) sqlInsert(command *jdb.Command) (string, []any) {
	from := command.GetFrom()
	if from == nil {
		return "", []any{}
	}

	table := from.Table
	columns := []string{}
	values := []string{}
	returns := []string{}
	args := []any{}
	data := et.Json{}
	for key, val := range command.New {
		field := from.GetField(key)
		switch field.TypeColumn {
		case jdb.TpColumn:
			if from.SourceField != nil && field.Name == from.SourceField.Name {
				continue
			}

			columns = append(columns, field.Name)
			val := quote(val)
			if field.TypeData == jdb.TypeDataObject {
				val = fmt.Sprintf(`%v::jsonb`, val)
			}
			values = append(values, strs.Format(`%v`, val))
			returns = append(returns, strs.Format("'%s', %s", key, key))
		case jdb.TpAtribute:
			val := quote(val)
			data.Set(key, val)
		}
	}

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
* @return string
**/
func (s *Postgres) sqlUpdate(command *jdb.Command) (string, []any) {
	args := []any{}
	from := command.GetFrom()
	if from == nil {
		return "", args
	}

	set := []string{}
	_data := ""
	for _, value := range command.Values {
		for key, field := range value {
			switch field.Column.TypeColumn {
			case jdb.TpColumn:
				if from.SourceField != nil && field.Column.Name == from.SourceField.Name {
					continue
				}
				val := field.ValueQuoted()
				set = append(set, strs.Format(`%s = %v`, key, val))
			case jdb.TpAtribute:
				val, tp := field.ValueToJSON()
				if len(fmt.Sprintf(`%v`, val)) == 0 {
					continue
				} else if fmt.Sprintf(`%v`, val) == "''" {
					continue
				} else if len(_data) == 0 {
					_data = fmt.Sprintf("COALESCE(%s, '{}')", from.SourceField.Name)
					_data = strs.Format("jsonb_set(%s,\n'{%s}', to_jsonb(%v::%s), true)", _data, key, val, tp)
				} else {
					_data = strs.Format("jsonb_set(\n%s,\n'{%s}', to_jsonb(%v::%s), true)", _data, key, val, tp)
				}
			}
		}
	}

	returns := []string{}
	if len(command.Returns) > 0 {
		for _, field := range command.Returns {
			if from.SourceField != nil && field.Column.Name == from.SourceField.Name {
				continue
			}
			if field.Column.TypeColumn != jdb.TpColumn {
				continue
			}
			name := field.Name
			returns = append(returns, strs.Format("'%s', %s", name, name))
		}
	} else {
		for _, field := range from.Model.Columns {
			if from.SourceField != nil && field.Name == from.SourceField.Name {
				continue
			}
			if field.TypeColumn != jdb.TpColumn {
				continue
			}
			name := field.Name
			returns = append(returns, strs.Format("'%s', %s", name, name))
		}
	}

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
* SqlDelete
* @param command *jdb.Command
* @return string
**/
func (s *Postgres) sqlDelete(command *jdb.Command) (string, []any) {
	args := []any{}
	from := command.GetFrom()
	if from == nil {
		return "", args
	}

	returns := []string{}
	if len(command.Returning) > 0 {
		for _, field := range command.Returning {
			if from.SourceField != nil && field.Name == from.SourceField.Name {
				continue
			}
			if field.TypeColumn != jdb.TpColumn {
				continue
			}
			name := field.Name
			returns = append(returns, strs.Format("'%s', %s", name, name))
		}
	} else {
		for _, field := range from.Model.Columns {
			if from.SourceField != nil && field.Name == from.SourceField.Name {
				continue
			}
			if field.TypeColumn != jdb.TpColumn {
				continue
			}
			name := field.Name
			returns = append(returns, strs.Format("'%s', %s", name, name))
		}
	}

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

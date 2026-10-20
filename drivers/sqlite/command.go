package sqlite

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
* jsonColumnTypes son los TypeData que se guardan como texto JSON y se insertan con json(?N).
**/
var jsonColumnTypes = []jdb.TypeData{jdb.TypeDataObject, jdb.TypeDataMultiSelect, jdb.TypeDataGeometry}

/**
* returningResult: expresión del RETURNING, un objeto JSON con las columnas (las pedidas en
* Returns(...) o todas) fijadas sobre el source, como `source || jsonb_build_object(...)` en postgres.
* @param from *jdb.QlFrom, requested []*jdb.Field
* @return string
**/
func returningResult(from *jdb.QlFrom, requested []*jdb.Field) string {
	isReturnable := func(name string, typeColumn jdb.TypeColumn) bool {
		if from.SourceField != nil && name == from.SourceField.Name {
			return false
		}
		return typeColumn == jdb.TpColumn
	}

	source := ""
	if from.SourceField != nil {
		source = quoteIdent(from.SourceField.Name)
	}

	pairs := []string{}
	if len(requested) > 0 {
		for _, field := range requested {
			if isReturnable(field.Name, field.TypeColumn) {
				pairs = append(pairs, jsonPair(field.Name, jsonExpr(quoteIdent(field.Name), field.TypeData), source != ""))
			}
		}
	} else {
		for _, column := range from.Model.Columns {
			if isReturnable(column.Name, column.TypeColumn) {
				pairs = append(pairs, jsonPair(column.Name, jsonExpr(quoteIdent(column.Name), column.TypeData), source != ""))
			}
		}
	}

	return strs.Format(`%s AS "result"`, jsonBuild(source, pairs))
}

/**
* sqlInsert: las columnas van como argumentos ligados y los atributos (declarados o no) se
* agrupan en el source, que viaja como un solo argumento JSON.
* @param command *jdb.Command
* @return string, []any
**/
func (s *Params) sqlInsert(command *jdb.Command) (string, []any) {
	from := command.GetFrom()
	if from == nil {
		return "", []any{}
	}

	columns := []string{}
	values := []string{}
	args := []any{}
	data := et.Json{}
	for key, val := range command.New {
		field := from.GetFieldOrAtribute(key)
		if field == nil {
			continue
		}

		switch field.TypeColumn {
		case jdb.TpColumn:
			if from.SourceField != nil && field.Name == from.SourceField.Name {
				continue
			}

			columns = append(columns, quoteIdent(field.Name))
			args = append(args, argValue(val))
			if slices.Contains(jsonColumnTypes, field.TypeData) {
				values = append(values, fmt.Sprintf(`json(?%d)`, len(args)))
			} else {
				values = append(values, fmt.Sprintf(`?%d`, len(args)))
			}
		case jdb.TpAtribute:
			data.Set(key, val)
		}
	}

	if from.SourceField != nil {
		columns = append(columns, quoteIdent(from.SourceField.Name))
		args = append(args, data.ToString())
		values = append(values, fmt.Sprintf(`json(?%d)`, len(args)))
	}

	result := strs.Format("INSERT INTO %s(\n%s)\nVALUES (%s)\nRETURNING\n%s;", tableName(from.Model), strings.Join(columns, ",\n"), strings.Join(values, ","), returningResult(from, command.Returning))

	return result, args
}

/**
* sqlUpdate: cada atributo se fija con json_set sobre el source (reemplazo de primer nivel,
* igual que `source || $1::jsonb` en postgres); ruta y valor van como argumentos ligados.
* @param command *jdb.Command
* @return string, []any
**/
func (s *Params) sqlUpdate(command *jdb.Command) (string, []any) {
	args := []any{}
	from := command.GetFrom()
	if from == nil {
		return "", args
	}

	set := []string{}
	attrs := []string{}
	for key, val := range command.New {
		field := from.GetFieldOrAtribute(key)
		if field == nil {
			continue
		}

		switch field.TypeColumn {
		case jdb.TpColumn:
			if from.SourceField != nil && field.Name == from.SourceField.Name {
				continue
			}

			args = append(args, argValue(val))
			if slices.Contains(jsonColumnTypes, field.TypeData) {
				set = append(set, fmt.Sprintf(`%s = json(?%d)`, quoteIdent(field.Name), len(args)))
			} else {
				set = append(set, fmt.Sprintf(`%s = ?%d`, quoteIdent(field.Name), len(args)))
			}
		case jdb.TpAtribute:
			args = append(args, fmt.Sprintf(`$."%s"`, strings.ReplaceAll(key, `"`, ``)), jsonArg(val))
			attrs = append(attrs, fmt.Sprintf(`?%d, json(?%d)`, len(args)-1, len(args)))
		}
	}

	if len(attrs) > 0 && from.SourceField != nil {
		column := quoteIdent(from.SourceField.Name)
		set = append(set, fmt.Sprintf("%s = json_set(COALESCE(%s, '{}'),\n%s)", column, column, strings.Join(attrs, ",\n")))
	}

	if len(set) == 0 {
		return "", args
	}

	result := strs.Format("UPDATE %s\nSET\n%s\nWHERE %s\nRETURNING\n%s;", s.tableAs(from), strings.Join(set, ",\n"), whereConditions(command.QlWhere), returningResult(from, command.Returning))

	return result, args
}

/**
* sqlDelete
* @param command *jdb.Command
* @return string, []any
**/
func (s *Params) sqlDelete(command *jdb.Command) (string, []any) {
	args := []any{}
	from := command.GetFrom()
	if from == nil {
		return "", args
	}

	result := strs.Format("DELETE FROM %s\nWHERE %s\nRETURNING\n%s;", s.tableAs(from), whereConditions(command.QlWhere), returningResult(from, command.Returning))

	return result, args
}

/**
* Command
* @param command *jdb.Command
* @return et.Items, error
**/
func (s *Params) Command(command *jdb.Command) (et.Items, error) {
	command.Sql = ""
	command.Args = []any{}
	switch command.Command {
	case jdb.Insert:
		command.Sql, command.Args = s.sqlInsert(command)
	case jdb.Update, jdb.Delete:
		if command.QlWhere == nil || len(command.QlWhere.Wheres) == 0 {
			return et.Items{}, fmt.Errorf(MSG_WHERE_REQUIRED)
		}

		if command.Command == jdb.Update {
			command.Sql, command.Args = s.sqlUpdate(command)
		} else {
			command.Sql, command.Args = s.sqlDelete(command)
		}
	default:
		return et.Items{}, fmt.Errorf(MSG_COMMAND_NOT_FOUND)
	}

	if command.Sql == "" {
		return et.Items{}, nil
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

	return normalizeResult(result), nil
}

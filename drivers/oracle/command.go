package oracle

import (
	"fmt"
	"strings"

	"github.com/celsiainternet/elvis/console"
	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/strs"
	"github.com/celsiainternet/jdb/jdb"
	goOra "github.com/sijms/go-ora/v2"
)

/**
* maxInList: Oracle admite hasta 1000 expresiones en un IN.
**/
const maxInList = 1000

/**
* returningResult: objeto JSON con las columnas (las pedidas en Returns(...) o todas) combinado con
* el source, como `source || jsonb_build_object(...)` en postgres. Oracle no tiene RETURNING de
* varias filas, así que esta expresión se lee con un SELECT por ROWID después de escribir.
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

	column := func(name string) string {
		return strs.Append(from.As, quoteIdent(name), ".")
	}

	source := ""
	if from.SourceField != nil {
		source = column(from.SourceField.Name)
	}

	pairs := []string{}
	if len(requested) > 0 {
		for _, field := range requested {
			if isReturnable(field.Name, field.TypeColumn) {
				pairs = append(pairs, jsonPair(field.Name, jsonExpr(column(field.Name), field.TypeData)))
			}
		}
	} else {
		for _, col := range from.Model.Columns {
			if isReturnable(col.Name, col.TypeColumn) {
				pairs = append(pairs, jsonPair(col.Name, jsonExpr(column(col.Name), col.TypeData)))
			}
		}
	}

	return strs.Format(`%s AS "result"`, jsonBuild(source, pairs))
}

/**
* sqlInsert: las columnas van como argumentos ligados y los atributos (declarados o no) se
* agrupan en el source, que viaja como un solo CLOB JSON. Devuelve el ROWID en un parámetro de
* salida que Command agrega al final.
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
			args = append(args, argValue(field.TypeData, val))
			values = append(values, fmt.Sprintf(`:%d`, len(args)))
		case jdb.TpAtribute:
			data.Set(key, val)
		}
	}

	if from.SourceField != nil {
		columns = append(columns, quoteIdent(from.SourceField.Name))
		args = append(args, goOra.Clob{String: data.ToString(), Valid: true})
		values = append(values, fmt.Sprintf(`:%d`, len(args)))
	}

	result := strs.Format("INSERT INTO %s(\n%s)\nVALUES (%s)\nRETURNING ROWIDTOCHAR(ROWID) INTO :%d", tableName(from.Model), strings.Join(columns, ",\n"), strings.Join(values, ","), len(args)+1)

	return result, args
}

/**
* sqlUpdate: los atributos se combinan con el source por JSON_MERGEPATCH (19c no tiene
* JSON_TRANSFORM): un valor null borra la clave y los objetos anidados se mezclan en vez de reemplazarse.
* @param command *jdb.Command, where string
* @return string, []any
**/
func (s *Params) sqlUpdate(command *jdb.Command, where string) (string, []any) {
	args := []any{}
	from := command.GetFrom()
	if from == nil {
		return "", args
	}

	set := []string{}
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

			args = append(args, argValue(field.TypeData, val))
			set = append(set, fmt.Sprintf(`%s = :%d`, quoteIdent(field.Name), len(args)))
		case jdb.TpAtribute:
			data.Set(key, val)
		}
	}

	if len(data) > 0 && from.SourceField != nil {
		args = append(args, goOra.Clob{String: data.ToString(), Valid: true})
		column := quoteIdent(from.SourceField.Name)
		set = append(set, fmt.Sprintf(`%s = JSON_MERGEPATCH(NVL(%s, '{}'), :%d RETURNING CLOB)`, column, column, len(args)))
	}

	if len(set) == 0 {
		return "", args
	}

	result := strs.Format("UPDATE %s\nSET\n%s\nWHERE %s", s.tableAs(from), strings.Join(set, ",\n"), where)

	return result, args
}

/**
* sqlDelete
* @param command *jdb.Command, where string
* @return string, []any
**/
func (s *Params) sqlDelete(command *jdb.Command, where string) (string, []any) {
	from := command.GetFrom()
	if from == nil {
		return "", []any{}
	}

	return strs.Format("DELETE FROM %s\nWHERE %s", s.tableAs(from), where), []any{}
}

/**
* sqlRowids: bloquea y lista las filas afectadas por el WHERE del comando.
* @param command *jdb.Command
* @return string
**/
func (s *Params) sqlRowids(command *jdb.Command) string {
	from := command.GetFrom()
	return strs.Format("SELECT ROWIDTOCHAR(%s) AS \"rid\"\nFROM %s\nWHERE %s\nFOR UPDATE", strs.Append(from.As, "ROWID", "."), s.tableAs(from), whereConditions(command.QlWhere))
}

/**
* rowidsWhere: condición por ROWID, en grupos de maxInList.
* @param from *jdb.QlFrom, rowids []string
* @return string
**/
func rowidsWhere(from *jdb.QlFrom, rowids []string) string {
	column := strs.Append(from.As, "ROWID", ".")
	groups := []string{}
	for i := 0; i < len(rowids); i += maxInList {
		end := min(i+maxInList, len(rowids))
		values := []string{}
		for _, rowid := range rowids[i:end] {
			values = append(values, fmt.Sprintf(`CHARTOROWID(%v)`, quote(rowid)))
		}
		groups = append(groups, fmt.Sprintf(`%s IN (%s)`, column, strings.Join(values, ", ")))
	}

	return strs.Format(`(%s)`, strings.Join(groups, "\nOR "))
}

/**
* sqlResult
* @param command *jdb.Command, where string
* @return string
**/
func (s *Params) sqlResult(command *jdb.Command, where string) string {
	from := command.GetFrom()
	return strs.Format("SELECT %s\nFROM %s\nWHERE %s", returningResult(from, command.Returning), s.tableAs(from), where)
}

/**
* execTx
* @param tx *jdb.Tx, query string, args ...any
* @return error
**/
func (s *Params) execTx(tx *jdb.Tx, query string, args ...any) error {
	var err error
	if tx != nil && tx.Tx != nil {
		_, err = tx.Tx.Exec(query, args...)
	} else {
		_, err = s.jdb.Db.Exec(query, args...)
	}

	return err
}

/**
* rowids
* @param command *jdb.Command
* @return []string, error
**/
func (s *Params) rowids(command *jdb.Command) ([]string, error) {
	sql := s.sqlRowids(command)
	if command.IsDebug {
		console.Debug(sql)
	}

	items, err := jdb.QueryTx(s.jdb, command.Tx(), sql)
	if err != nil {
		return nil, err
	}

	result := []string{}
	for _, item := range items.Result {
		result = append(result, item.Str("rid"))
	}

	return result, nil
}

/**
* Command: Oracle no devuelve filas desde INSERT/UPDATE/DELETE, así que se trabaja por ROWID:
* INSERT devuelve el suyo con RETURNING INTO; UPDATE y DELETE primero bloquean (FOR UPDATE) las filas
* del WHERE y luego leen el resultado (después de UPDATE, antes de DELETE).
* @param command *jdb.Command
* @return et.Items, error
**/
func (s *Params) Command(command *jdb.Command) (et.Items, error) {
	command.Sql = ""
	command.Args = []any{}
	from := command.GetFrom()
	if from == nil {
		return et.Items{}, fmt.Errorf(MSG_COMMAND_NOT_FOUND)
	}

	debug := func() {
		if command.IsDebug {
			console.Debug(et.Json{"sql": command.Sql, "args": command.Args}.ToString())
		}
	}

	fail := func(err error) (et.Items, error) {
		console.Error(err)
		console.Debug(command.Sql)
		console.Debug(command.Args)
		return et.Items{}, err
	}

	switch command.Command {
	case jdb.Insert:
		command.Sql, command.Args = s.sqlInsert(command)
		debug()

		var rowid string
		args := append(append([]any{}, command.Args...), goOra.Out{Dest: &rowid, Size: 100})
		if err := s.execTx(command.Tx(), command.Sql, args...); err != nil {
			return fail(err)
		}

		if rowid == "" {
			return fail(fmt.Errorf(MSG_ROWID_NOT_FOUND))
		}

		return s.result(command, rowidsWhere(from, []string{rowid}))
	case jdb.Update, jdb.Delete:
		if command.QlWhere == nil || len(command.QlWhere.Wheres) == 0 {
			return et.Items{}, fmt.Errorf(MSG_WHERE_REQUIRED)
		}

		rowids, err := s.rowids(command)
		if err != nil {
			return fail(err)
		}

		if len(rowids) == 0 {
			return et.Items{}, nil
		}

		where := rowidsWhere(from, rowids)
		if command.Command == jdb.Update {
			command.Sql, command.Args = s.sqlUpdate(command, where)
			if command.Sql == "" {
				return et.Items{}, nil
			}

			debug()
			if err := s.execTx(command.Tx(), command.Sql, command.Args...); err != nil {
				return fail(err)
			}

			return s.result(command, where)
		}

		result, err := s.result(command, where)
		if err != nil {
			return et.Items{}, err
		}

		command.Sql, command.Args = s.sqlDelete(command, where)
		debug()
		if err := s.execTx(command.Tx(), command.Sql, command.Args...); err != nil {
			return fail(err)
		}

		return result, nil
	default:
		return et.Items{}, fmt.Errorf(MSG_COMMAND_NOT_FOUND)
	}
}

/**
* result: lee las filas escritas como objetos JSON.
* @param command *jdb.Command, where string
* @return et.Items, error
**/
func (s *Params) result(command *jdb.Command, where string) (et.Items, error) {
	sql := s.sqlResult(command, where)
	if command.IsDebug {
		console.Debug(sql)
	}

	items, err := jdb.QueryTx(s.jdb, command.Tx(), sql)
	if err != nil {
		return et.Items{}, err
	}

	return normalizeResult(items), nil
}

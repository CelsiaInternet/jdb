package postgres

import (
	"fmt"
	"strings"

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

	args := []any{}
	columns := []string{}
	values := []string{}
	returns := []string{}
	_data := et.Json{}
	for _, val := range command.Values {
		for key, field := range val {
			switch field.Column.TypeColumn {
			case jdb.TpColumn:
				if from.SourceField != nil && field.Column.Name == from.SourceField.Name {
					continue
				}
				columns = append(columns, key)
				val := field.ValueQuoted()
				values = append(values, strs.Format(`%v`, val))
				returns = append(returns, strs.Format("'%s', %s", key, key))
			case jdb.TpAtribute:
				_data.Set(key, field.Value)
			}
		}
	}

	table := tableName(from.Model)
	result := "INSERT INTO %s(\n%s)\nVALUES (%s)\nRETURNING\njsonb_build_object(%s) AS result;"
	if from.SourceField != nil {
		column := from.SourceField.Name
		columns = append(columns, column)
		arg := strs.Format(`%v`, _data.ToString())
		args = append(args, arg)
		values = append(values, fmt.Sprintf(`$%d::jsonb`, len(args)))

		result = "INSERT INTO %s(\n%s)\nVALUES (%s)\nRETURNING\n%s || jsonb_build_object(%s) AS result;"
		result = strs.Format(result, table, strings.Join(columns, ",\n"), strings.Join(values, ","), from.SourceField.Name, strings.Join(returns, ","))
	} else {
		result = strs.Format(result, table, strings.Join(columns, ",\n"), strings.Join(values, ","), strings.Join(returns, ","))
	}

	return result, args
}

package postgres

import (
	"strings"

	"github.com/celsiainternet/elvis/strs"
	jdb "github.com/celsiainternet/jdb/jdb"
)

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
	result := "DELETE FROM %s\nWHERE %s\nRETURNING\njsonb_build_object(%s) AS result;"
	if from.SourceField != nil {
		result = "DELETE FROM %s\nWHERE %s\nRETURNING\n%s || jsonb_build_object(%s) AS result;"
		result = strs.Format(result, table, where, from.SourceField.Name, strings.Join(returns, ","))
	} else {
		result = strs.Format(result, table, where, strings.Join(returns, ","))
	}
	return result, args
}

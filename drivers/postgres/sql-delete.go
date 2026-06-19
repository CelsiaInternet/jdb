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
	for _, val := range command.Values {
		for key, field := range val {
			switch field.Column.TypeColumn {
			case jdb.TpColumn:
				if field.Column.Name == from.SourceField.Name {
					continue
				}
				returns = append(returns, strs.Format("'%s', %s", key, key))
			}
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

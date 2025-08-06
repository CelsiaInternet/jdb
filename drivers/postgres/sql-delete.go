package postgres

import (
	"github.com/celsiainternet/elvis/strs"
	jdb "github.com/celsiainternet/jdb/jdb"
)

/**
* SqlDelete
* @param command *jdb.Command
* @return string
**/
func (s *Postgres) sqlDelete(command *jdb.Command) string {
	from := command.From
	where := whereConditions(command.QlWhere)
	objects := s.sqlObject(from.GetFrom())
	returns := strs.Format("%s AS result", objects)
	if len(command.Returns) > 0 {
		returns = ""
		for _, fld := range command.Returns {
			returns = strs.Append(returns, fld.Name, ", ")
		}
	}
	result := "DELETE FROM %s\nWHERE %s\nRETURNING\n%s;"
	return strs.Format(result, tableName(from), where, returns)
}

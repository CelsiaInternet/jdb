package postgres

import (
	"fmt"
	"strings"

	"github.com/celsiainternet/elvis/console"
	"github.com/celsiainternet/elvis/strs"
	jdb "github.com/celsiainternet/jdb/jdb"
)

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
	returns := []string{}
	_data := ""
	for _, value := range command.Values {
		for key, field := range value {
			switch field.Column.TypeColumn {
			case jdb.TpColumn:
				if field.Column.Name == from.SourceField.Name {
					continue
				}
				arg := strs.Format(`%v`, field.Value)
				args = append(args, arg)
				set = append(set, strs.Format(`%s = $%d`, key, len(args)))
				returns = append(returns, strs.Format("'%s', %s", key, key))
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

	console.Debug(result)

	return result, args
}

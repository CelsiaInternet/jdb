package mysql

import (
	"github.com/celsiainternet/elvis/strs"
	jdb "github.com/celsiainternet/jdb/jdb"
)

/**
* sqlHaving
* @param ql *jdb.Ql
* @return string
**/
func (s *Mysql) sqlHaving(ql *jdb.Ql) string {
	result := ""
	havings := ql.Havings
	where := whereConditions(havings.QlWhere)
	if where == "" {
		return result
	}

	result = strs.Format("HAVING %s", where)

	return result
}

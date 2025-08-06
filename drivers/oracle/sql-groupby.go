package oracle

import (
	"github.com/celsiainternet/elvis/strs"
	jdb "github.com/celsiainternet/jdb/jdb"
)

/**
* sqlGroupBy
* @param ql *jdb.Ql
* @return string
**/
func (s *Oracle) sqlGroupBy(ql *jdb.Ql) string {
	result := ""
	columns := s.sqlColumns(ql.Groups)
	if len(columns) == 0 {
		return result
	}

	result = strs.Format("GROUP BY %s", columns)

	return result
}

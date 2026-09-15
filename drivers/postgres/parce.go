package postgres

import (
	"fmt"
	"strings"
)

/**
* SQLDDL return a sql string with the args
* @param sql string
* @param args ...any
* @return string
**/
func sqlDDL(sql string, args ...any) string {
	sql = strings.TrimSpace(sql)

	for i, arg := range args {
		old := fmt.Sprintf(`$%d`, i+1)
		new := fmt.Sprintf(`%v`, arg)
		sql = strings.ReplaceAll(sql, old, new)
	}

	return sql
}

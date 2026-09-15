package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/strs"
	jdb "github.com/celsiainternet/jdb/jdb"
)

/**
* sqlQuote return a sql cuote string
* @param sql string
* @return string
**/
func sqlQuote(sql string) string {
	sql = strings.TrimSpace(sql)

	result := strs.Replace(sql, `'`, `"`)
	result = strs.Trim(result)

	return result
}

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

/**
* SQLParse return a sql string with the args
* @param sql string
* @param args ...any
* @return string
**/
func sqlParse(sql string, args ...any) string {
	for i := range args {
		old := fmt.Sprintf(`$%d`, i+1)
		new := fmt.Sprintf(`{$%d}`, i+1)
		sql = strings.ReplaceAll(sql, old, new)
	}

	for i, arg := range args {
		old := fmt.Sprintf(`{$%d}`, i+1)
		new := fmt.Sprintf(`%v`, Quote(arg))
		sql = strings.ReplaceAll(sql, old, new)
	}

	return sql
}

/**
* rowsToItem return a item from a sql rows
* @param rows *sql.Rows
* @return et.Item
**/
func rowsToItem(rows *sql.Rows) et.Item {
	result := jdb.RowsToItems(rows)
	if result.Count == 0 {
		return et.Item{}
	}

	return result.First()
}

/**
* RowsToSourceItems return a items from a sql rows and source field
* @param rows *sql.Rows, source string
* @return et.Items
**/
func rowsToSourceItems(rows *sql.Rows, source string) et.Items {
	var result = et.Items{Result: []et.Json{}}
	for rows.Next() {
		var item et.Json
		item.ScanRows(rows)

		result.Ok = true
		result.Count++
		if item[source] == nil {
			result.Result = append(result.Result, item)
		} else {
			result.Result = append(result.Result, item.Json(source))
		}
	}

	return result
}

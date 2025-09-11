package jdb

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/celsiainternet/elvis/console"
	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/strs"
)

/**
* SQLQuote return a sql cuote string
* @param sql string
* @return string
**/
func SQLQuote(sql string) string {
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
func SQLDDL(sql string, args ...any) string {
	sql = strings.TrimSpace(sql)

	for i, arg := range args {
		old := strs.Format(`$%d`, i+1)
		new := strs.Format(`%v`, arg)
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
func SQLParse(sql string, args ...any) string {
	for i := range args {
		old := strs.Format(`$%d`, i+1)
		new := strs.Format(`{$%d}`, i+1)
		sql = strings.ReplaceAll(sql, old, new)
	}

	for i, arg := range args {
		old := strs.Format(`{$%d}`, i+1)
		new := strs.Format(`%v`, Quote(arg))
		sql = strings.ReplaceAll(sql, old, new)
	}

	return sql
}

/**
* RowsToItems return a items from a sql rows
* @param rows *sql.Rows
* @return et.Items
**/
func RowsToItems(rows *sql.Rows) et.Items {
	var result = et.Items{Result: []et.Json{}}
	for rows.Next() {
		var item et.Json
		item.ScanRows(rows)

		result.Ok = true
		result.Count++
		result.Result = append(result.Result, item)
	}

	return result
}

/**
* RowsToItem return a item from a sql rows
* @param rows *sql.Rows
* @return et.Item
**/
func RowsToItem(rows *sql.Rows) et.Item {
	result := RowsToItems(rows)
	if result.Count == 0 {
		return et.Item{}
	}

	return result.First()
}

/**
* RowsToSourceItem return a items from a sql rows and source field
* @param rows *sql.Rows, source string
* @return et.Items
**/
func RowsToSourceItem(rows *sql.Rows, source string) et.Items {
	var result = et.Items{Result: []et.Json{}}
	for rows.Next() {
		var item et.Json
		item.ScanRows(rows)

		result.Ok = true
		result.Count++
		result.Result = append(result.Result, item.Json(source))
	}

	return result
}

/**
* JsonQuote return a json quote string
* @param val interface{}
* @return interface{}
**/
func JsonQuote(val interface{}) interface{} {
	f := `'%v'`
	switch v := val.(type) {
	case string:
		v = strs.Format(`"%s"`, v)
		return strs.Format(f, v)
	case int:
		return strs.Format(f, v)
	case float64:
		return strs.Format(f, v)
	case float32:
		return strs.Format(f, v)
	case int16:
		return strs.Format(f, v)
	case int32:
		return strs.Format(f, v)
	case int64:
		return strs.Format(f, v)
	case bool:
		return strs.Format(f, v)
	case time.Time:
		return strs.Format(f, v.Format("2006-01-02 15:04:05"))
	case et.Json:
		return strs.Format(f, v.ToString())
	case map[string]interface{}:
		return strs.Format(f, et.Json(v).ToString())
	case []string:
		var r string
		for _, s := range v {
			r = strs.Append(r, strs.Format(`"%s"`, s), ", ")
		}
		r = strs.Format(`[%s]`, r)
		return strs.Format(f, r)
	case []interface{}:
		var r string
		for _, _v := range v {
			q := JsonQuote(_v)
			r = strs.Append(r, strs.Format(`%v`, q), ", ")
		}
		r = strs.Format(`[%s]`, r)
		return strs.Format(f, r)
	case []uint8:
		return strs.Format(f, string(v))
	case nil:
		return strs.Format(`%s`, "NULL")
	default:
		console.Alert(fmt.Sprintf("Not quoted type:%v value:%v", reflect.TypeOf(v), v))
		return val
	}
}

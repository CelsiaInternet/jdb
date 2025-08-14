package jdb

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/logs"
	"github.com/celsiainternet/elvis/strs"
)

var quotedChar = `'`

/**
* SetQuotedChar
* @param char string
**/
func SetQuotedChar(char string) {
	quotedChar = strs.Format(`%s`, char)
}

/**
* quote
* @param str string
* @return string
**/
func quote(str string) string {
	result := strconv.Quote(str)
	if quotedChar == `"` {
		return result
	}

	return strings.ReplaceAll(result, `"`, `'`)
}

/**
* Quote
* @param val interface{}
* @return any
**/
func Quote(val interface{}) any {
	fmt := `'%s'`
	if quotedChar == `"` {
		fmt = `"%s"`
	}
	switch v := val.(type) {
	case string:
		return quote(v)
	case int:
		return v
	case float64:
		return v
	case float32:
		return v
	case int16:
		return v
	case int32:
		return v
	case int64:
		return v
	case bool:
		return v
	case time.Time:
		return strs.Format(fmt, v.Format("2006-01-02 15:04:05"))
	case et.Json:
		return strs.Format(fmt, v.ToString())
	case map[string]interface{}:
		return strs.Format(fmt, et.Json(v).ToString())
	case []et.Json, []string, []interface{}, []map[string]interface{}:
		bt, err := json.Marshal(v)
		if err != nil {
			logs.Errorf("Quote type:%v, value:%v, error marshalling array: %v", reflect.TypeOf(v), v, err)
			return strs.Format(fmt, `[]`)
		}
		return strs.Format(fmt, string(bt))
	case []uint8:
		return strs.Format(fmt, string(v))
	case nil:
		return strs.Format(`%s`, "NULL")
	default:
		logs.Errorf("Quote type:%v, value:%v", reflect.TypeOf(v), v)
		return val
	}
}

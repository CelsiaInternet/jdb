package jdb

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/celsiainternet/elvis/console"
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
	case []string:
		bt, err := json.Marshal(v)
		if err != nil {
			logs.Errorf("Quote", "type:%v, value:%v, error marshalling array: %v", reflect.TypeOf(v), v, err)
			return strs.Format(fmt, `[]`)
		}
		return strs.Format(fmt, string(bt))
	case et.Json:
		return strs.Format(fmt, v.ToString())
	case map[string]interface{}:
		return strs.Format(fmt, et.Json(v).ToString())
	case []et.Json, []interface{}, []map[string]interface{}:
		bt, err := json.Marshal(v)
		if err != nil {
			logs.Errorf("Quote", "type:%v, value:%v, error marshalling array: %v", reflect.TypeOf(v), v, err)
			return strs.Format(fmt, `[]`)
		}
		return strs.Format(fmt, string(bt))
	case []uint8:
		return strs.Format(fmt, string(v))
	case nil:
		return strs.Format(`%s`, "NULL")
	default:
		logs.Errorf("Quote", "type:%v, value:%v", reflect.TypeOf(v), v)
		return val
	}
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
		v = fmt.Sprintf(`"%s"`, v)
		return fmt.Sprintf(f, v)
	case int:
		return fmt.Sprintf(f, v)
	case float64:
		return fmt.Sprintf(f, v)
	case float32:
		return fmt.Sprintf(f, v)
	case int16:
		return fmt.Sprintf(f, v)
	case int32:
		return fmt.Sprintf(f, v)
	case int64:
		return fmt.Sprintf(f, v)
	case bool:
		return fmt.Sprintf(f, v)
	case time.Time:
		return fmt.Sprintf(f, v.Format("2006-01-02 15:04:05"))
	case et.Json:
		return fmt.Sprintf(f, v.ToString())
	case map[string]interface{}:
		return fmt.Sprintf(f, et.Json(v).ToString())
	case []string:
		var r string
		for _, s := range v {
			r = strs.Append(r, fmt.Sprintf(`"%s"`, s), ", ")
		}
		r = fmt.Sprintf(`[%s]`, r)
		return fmt.Sprintf(f, r)
	case []et.Json, []interface{}, []map[string]interface{}:
		bt, err := json.Marshal(v)
		if err != nil {
			logs.Errorf("JsonQuote", "type:%v, value:%v, error marshalling array: %v", reflect.TypeOf(v), v, err)
			return strs.Format(f, `[]`)
		}
		return strs.Format(f, string(bt))
	case []uint8:
		return fmt.Sprintf(f, string(v))
	case nil:
		return fmt.Sprintf(`%s`, "NULL")
	default:
		console.Alert(fmt.Sprintf("Not quoted type:%v value:%v", reflect.TypeOf(v), v))
		return val
	}
}

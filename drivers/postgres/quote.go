package postgres

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/celsiainternet/elvis/console"
	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/logs"
	"github.com/celsiainternet/elvis/strs"
	jdb "github.com/celsiainternet/jdb/jdb"
)

/**
* Quote renders a Go value as a PostgreSQL SQL literal (or, for jdb.Value's
* Field/Agregation/Calc variants, a raw SQL expression). It is the single
* place responsible for translating both plain Go values and jdb's tagged
* jdb.Value wrapper (as carried by QlCondition.Value and Agregation.Value)
* into valid Postgres SQL text.
* @param val interface{}
* @return any
**/
func quote(val interface{}) any {
	format := `'%s'`
	switch v := val.(type) {
	case string:
		v = EscapeJSON(v)
		return fmt.Sprintf(format, v)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return v
	case bool:
		return v
	case time.Time:
		return fmt.Sprintf(format, v.Format("2006-01-02 15:04:05"))
	case jdb.Value:
		return quoteValue(v)
	case *jdb.Value:
		return quoteValue(*v)
	case et.Json:
		return strs.Format(format, v.ToString())
	case map[string]interface{}:
		return strs.Format(format, et.Json(v).ToString())
	case []string, []int, []int8, []int16, []int32, []int64, []uint, []uint16, []uint32, []uint64, []float32, []float64, []et.Json, []interface{}, []map[string]interface{}:
		bt, err := json.Marshal(v)
		if err != nil {
			logs.Errorf("Quote", "type:%v, value:%v, error marshalling array: %v", reflect.TypeOf(v), v, err)
			return strs.Format(format, `[]`)
		}
		return strs.Format(format, string(bt))
	case []uint8:
		b := []byte(val.([]uint8))
		return fmt.Sprintf("'\\x%s'", hex.EncodeToString(b))
	case nil:
		return fmt.Sprintf(`%s`, "NULL")
	default:
		logs.Errorf("Quote", "type:%v, value:%v", reflect.TypeOf(v), v)
		return val
	}
}

/**
* quoteValue renders a jdb.Value - the {Type, Value} wrapper QlCondition.Value
* and Agregation.Value carry - covering every jdb.ValueType variant:
*  - ValueTypeString/Number/DateTime/Boolean/Json/JsonArray/Array/Binary/Null
*    all wrap a plain Go value that quote() already knows how to render, so
*    they delegate back to quote() on the unwrapped value;
*  - ValueTypeField is not a literal at all - it names another column, so it
*    renders as a raw SQL column reference via asField;
*  - ValueTypeAgregation renders its aggregate expression via asAgregation;
*  - ValueTypeCalc carries a raw SQL expression string and is embedded as-is,
*    unquoted.
* @param v jdb.Value
* @return any
**/
func quoteValue(v jdb.Value) any {
	switch v.Type {
	case jdb.ValueTypeField:
		switch f := v.Value.(type) {
		case *jdb.Field:
			return asField(*f)
		case jdb.Field:
			return asField(f)
		default:
			return quote(v.Value)
		}
	case jdb.ValueTypeAgregation:
		switch a := v.Value.(type) {
		case *jdb.Agregation:
			return asAgregation(a)
		case jdb.Agregation:
			return asAgregation(&a)
		default:
			return quote(v.Value)
		}
	case jdb.ValueTypeCalc:
		return fmt.Sprintf(`%v`, v.Value)
	case jdb.ValueTypeString,
		jdb.ValueTypeNumber,
		jdb.ValueTypeDateTime,
		jdb.ValueTypeBoolean,
		jdb.ValueTypeJson,
		jdb.ValueTypeJsonArray,
		jdb.ValueTypeArray,
		jdb.ValueTypeBinary,
		jdb.ValueTypeNull:
		return quote(v.Value)
	default:
		logs.Errorf("Quote", "unhandled jdb.ValueType:%v value:%v", v.Type, v.Value)
		return quote(v.Value)
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
		v = strs.Replace(v, `'`, `''`)
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

/**
* EscapeJSON
* @param val string
* @return string
**/
func EscapeJSON(val string) string {
	val = strings.ReplaceAll(val, `'`, `''`)
	return val
}

/**
* Normalize
* @param v any
* @return any
**/
func Normalize(v any) any {
	switch x := v.(type) {

	case map[string]any:
		for k, val := range x {
			x[k] = Normalize(val)

			if s, ok := x[k].(string); ok {
				var j any

				if json.Unmarshal([]byte(s), &j) == nil {
					x[k] = j
				}
			}
		}

	case []any:
		for i, val := range x {
			x[i] = Normalize(val)
		}
	}

	return v
}

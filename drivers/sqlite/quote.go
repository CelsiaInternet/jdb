package sqlite

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/logs"
	"github.com/celsiainternet/elvis/strs"
	jdb "github.com/celsiainternet/jdb/jdb"
)

/**
* quoteIdent: identificador entre comillas dobles (SQLite reserva palabras como index u order).
* @param name string
* @return string
**/
func quoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

/**
* jsonPath: literal de ruta JSON '$."key"' para una clave de primer nivel.
* @param key string
* @return string
**/
func jsonPath(key string) string {
	key = strings.ReplaceAll(key, `"`, ``)
	key = strings.ReplaceAll(key, `'`, `''`)
	return strs.Format(`'$."%s"'`, key)
}

/**
* quote: renderiza un valor Go como literal SQLite (o, para los jdb.Value de tipo
* Field/Agregation/Calc, como expresión SQL cruda).
* @param val interface{}
* @return any
**/
func quote(val interface{}) any {
	format := `'%s'`
	switch v := val.(type) {
	case string:
		return fmt.Sprintf(format, strings.ReplaceAll(v, `'`, `''`))
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return v
	case bool:
		if v {
			return 1
		}
		return 0
	case time.Time:
		return fmt.Sprintf(format, v.Format("2006-01-02 15:04:05"))
	case jdb.Value:
		return quoteValue(v)
	case *jdb.Value:
		return quoteValue(*v)
	case et.Json:
		return quote(v.ToString())
	case jdb.Agregation:
		return aliasAsAgregation(&v)
	case *jdb.Agregation:
		return aliasAsAgregation(v)
	case map[string]interface{}:
		return quote(et.Json(v).ToString())
	case []string, []int, []int8, []int16, []int32, []int64, []uint, []uint16, []uint32, []uint64, []float32, []float64, []et.Json, []interface{}, []map[string]interface{}:
		bt, err := json.Marshal(v)
		if err != nil {
			logs.Errorf("Quote", "type:%v, value:%v, error marshalling array: %v", reflect.TypeOf(v), v, err)
			return `'[]'`
		}
		return quote(string(bt))
	case []uint8:
		return fmt.Sprintf("X'%s'", hex.EncodeToString(v))
	case nil:
		return "NULL"
	default:
		logs.Errorf("Quote", "type:%v, value:%v", reflect.TypeOf(v), v)
		return val
	}
}

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
	default:
		return quote(v.Value)
	}
}

/**
* argValue: valor que se pasa como argumento ligado; los mapas y slices viajan como texto JSON.
* @param val interface{}
* @return any
**/
func argValue(val interface{}) any {
	switch v := val.(type) {
	case nil, string, bool, []byte, time.Time,
		int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return v
	case et.Json:
		return v.ToString()
	default:
		bt, err := json.Marshal(v)
		if err != nil {
			logs.Errorf("argValue", "type:%v, value:%v, error marshalling: %v", reflect.TypeOf(v), v, err)
			return nil
		}
		return string(bt)
	}
}

/**
* jsonArg: valor codificado como JSON, para ligarlo dentro de json(?N) y conservar su tipo.
* @param val interface{}
* @return string
**/
func jsonArg(val interface{}) string {
	if v, ok := val.(time.Time); ok {
		val = v.Format(time.RFC3339Nano)
	}

	bt, err := json.Marshal(val)
	if err != nil {
		logs.Errorf("jsonArg", "type:%v, value:%v, error marshalling: %v", reflect.TypeOf(val), val, err)
		return "null"
	}

	return string(bt)
}

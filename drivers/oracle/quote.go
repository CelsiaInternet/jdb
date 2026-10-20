package oracle

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
	"github.com/celsiainternet/jdb/jdb"
	goOra "github.com/sijms/go-ora/v2"
)

/**
* quoteIdent: identificador entre comillas dobles; sin ellas Oracle pasa todo a mayúsculas y
* rechaza nombres que empiezan por "_" (como _id o _data).
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
* quote: renderiza un valor Go como literal Oracle (o, para los jdb.Value de tipo
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
		return fmt.Sprintf(`TIMESTAMP '%s'`, v.Format("2006-01-02 15:04:05"))
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
		return fmt.Sprintf("HEXTORAW('%s')", hex.EncodeToString(v))
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
* dateLayouts: formatos de texto que se aceptan para columnas TIMESTAMP (el primero es el de elvis/timezone).
**/
var dateLayouts = []string{"2006/01/02 15:04:05", time.RFC3339Nano, "2006-01-02 15:04:05", "2006-01-02T15:04:05", "2006-01-02"}

/**
* argValue: valor que se liga en INSERT/UPDATE según el tipo de la columna: booleanos como 1/0,
* fechas en texto como time.Time, y JSON o texto largo como CLOB.
* @param typeData jdb.TypeData, val interface{}
* @return any
**/
func argValue(typeData jdb.TypeData, val interface{}) any {
	if val == nil {
		return nil
	}

	switch typeData {
	case jdb.TypeDataCheckbox:
		switch v := val.(type) {
		case bool:
			if v {
				return 1
			}
			return 0
		}
	case jdb.TypeDataDateTime:
		if str, ok := val.(string); ok {
			for _, layout := range dateLayouts {
				if t, err := time.Parse(layout, str); err == nil {
					return t
				}
			}
		}
	case jdb.TypeDataObject, jdb.TypeDataMultiSelect, jdb.TypeDataGeometry:
		if str, ok := val.(string); ok {
			return goOra.Clob{String: str, Valid: true}
		}
		return goOra.Clob{String: jsonArg(val), Valid: true}
	case jdb.TypeDataMemo, jdb.TypeDataFullText, jdb.TypeDataFilesMedia, jdb.TypeDataUrl, jdb.TypeDataAddress:
		if str, ok := val.(string); ok {
			return goOra.Clob{String: str, Valid: true}
		}
	}

	switch v := val.(type) {
	case string, []byte, time.Time,
		int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return v
	case bool:
		if v {
			return 1
		}
		return 0
	default:
		return jsonArg(v)
	}
}

/**
* jsonArg: valor codificado como JSON.
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

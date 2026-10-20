package oracle

import (
	"encoding/json"

	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/jdb/jdb"
)

/**
* normalizeResult: go-ora devuelve el CLOB de la columna result como texto, y
* jdb.RowsToItems solo desempaqueta []byte; aquí se convierte cada fila en su objeto.
* @param items et.Items
* @return et.Items
**/
func normalizeResult(items et.Items) et.Items {
	for i, item := range items.Result {
		if len(item) != 1 {
			continue
		}

		var raw []byte
		switch v := item["result"].(type) {
		case string:
			raw = []byte(v)
		case []byte:
			raw = v
		default:
			continue
		}

		var result et.Json
		if err := json.Unmarshal(raw, &result); err == nil {
			items.Result[i] = result
		}
	}

	return items
}

/**
* selectTypes: tipo de dato de cada columna proyectada, por nombre de salida.
* @param ql *jdb.Ql
* @return map[string]jdb.TypeData
**/
func selectTypes(ql *jdb.Ql) map[string]jdb.TypeData {
	result := map[string]jdb.TypeData{}
	add := func(field *jdb.Field) {
		name := field.As
		if name == "" {
			name = field.Name
		}
		result[name] = field.TypeData
	}

	if len(ql.Selects) == 0 && len(ql.Froms.Froms) > 0 {
		for _, field := range ql.Froms.Froms[0].GetColumnsFields() {
			add(field)
		}
		return result
	}

	for _, sel := range ql.Selects {
		switch v := sel.(type) {
		case *jdb.Field:
			add(v)
		case jdb.Field:
			add(&v)
		}
	}

	return result
}

/**
* normalizeColumns: devuelve los booleanos como bool y las columnas JSON como objeto,
* igual que llegan desde postgres.
* @param items et.Items, types map[string]jdb.TypeData
* @return et.Items
**/
func normalizeColumns(items et.Items, types map[string]jdb.TypeData) et.Items {
	for _, item := range items.Result {
		for name, tp := range types {
			val, ok := item[name]
			if !ok || val == nil {
				continue
			}

			switch tp {
			case jdb.TypeDataCheckbox:
				item[name] = toInt(val) == 1
			case jdb.TypeDataObject, jdb.TypeDataMultiSelect, jdb.TypeDataGeometry:
				str, ok := val.(string)
				if !ok {
					continue
				}
				var parsed interface{}
				if err := json.Unmarshal([]byte(str), &parsed); err == nil {
					item[name] = parsed
				}
			}
		}
	}

	return items
}

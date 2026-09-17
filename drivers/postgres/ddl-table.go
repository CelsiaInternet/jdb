package postgres

import (
	"fmt"

	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/strs"
	"github.com/celsiainternet/elvis/utility"
	jdb "github.com/celsiainternet/jdb/jdb"
)

/**
* typeData
* @param tp jdb.TypeData
* @return interface{}
**/
func (s *Params) typeData(tp jdb.TypeData) interface{} {
	switch tp {
	case jdb.TypeDataText:
		return "VARCHAR(250)"
	case jdb.TypeDataMemo:
		return "TEXT"
	case jdb.TypeDataShortText:
		return "VARCHAR(80)"
	case jdb.TypeDataKey:
		return "VARCHAR(80)"
	case jdb.TypeDataNumber:
		return "DECIMAL(18,2)"
	case jdb.TypeDataInt:
		return "BIGINT"
	case jdb.TypeDataPrecision:
		return "DOUBLE PRECISION"
	case jdb.TypeDataDateTime:
		return "TIMESTAMP"
	case jdb.TypeDataCheckbox:
		return "BOOLEAN"
	case jdb.TypeDataBytes:
		return "BYTEA"
	case jdb.TypeDataObject:
		return "JSONB"
	case jdb.TypeDataSelect:
		return "VARCHAR(250)"
	case jdb.TypeDataMultiSelect:
		return "JSONB"
	case jdb.TypeDataGeometry:
		return "JSONB"
	case jdb.TypeDataFullText:
		return "TSVECTOR"
	case jdb.TypeDataState:
		return "VARCHAR(80)"
	case jdb.TypeDataUser:
		return "VARCHAR(250)"
	case jdb.TypeDataFilesMedia:
		return "TEXT"
	case jdb.TypeDataUrl:
		return "TEXT"
	case jdb.TypeDataEmail:
		return "VARCHAR(250)"
	case jdb.TypeDataPhone:
		return "VARCHAR(250)"
	case jdb.TypeDataAddress:
		return "TEXT"
	case jdb.TypeDataRelation:
		return "VARCHAR(80)"
	case jdb.TypeDataRollup:
		return "VARCHAR(80)"
	default:
		return "VARCHAR(250)"
	}
}

/**
* defaultValue
* @param tp jdb.TypeData
* @return interface{}
**/
func (s *Params) defaultValue(tp jdb.TypeData) interface{} {
	switch tp {
	case jdb.TypeDataNumber:
		return 0.0
	case jdb.TypeDataInt:
		return 0
	case jdb.TypeDataPrecision:
		return 0.0
	case jdb.TypeDataDateTime:
		return "NOW()"
	case jdb.TypeDataCheckbox:
		return quote(false)
	case jdb.TypeDataBytes:
		return quote("")
	case jdb.TypeDataObject:
		return quote(et.Json{})
	case jdb.TypeDataSelect:
		return quote("")
	case jdb.TypeDataMultiSelect:
		return quote([]et.Json{})
	case jdb.TypeDataGeometry:
		return quote(et.Json{
			"type":        "Point",
			"coordinates": []float64{0, 0},
		})
	case jdb.TypeDataFullText:
		return quote("")
	case jdb.TypeDataState:
		return quote(utility.ACTIVE)
	case jdb.TypeDataUser:
		return quote("")
	case jdb.TypeDataFilesMedia:
		return quote("")
	case jdb.TypeDataUrl:
		return quote("")
	case jdb.TypeDataEmail:
		return quote("")
	case jdb.TypeDataPhone:
		return quote("")
	case jdb.TypeDataAddress:
		return quote("")
	case jdb.TypeDataRelation:
		return quote("")
	case jdb.TypeDataRollup:
		return quote("")
	default:
		return quote("")
	}
}

/**
* tableName
* @param model *jdb.Model
* @return string
**/
func tableName(model *jdb.Model) string {
	return fmt.Sprintf(`%s.%s`, model.Schema, model.Table)
}

/**
* existTable
* @param schema, name string
* @return bool, error
**/
func (s *Params) existTable(schema, name string) (bool, error) {
	sql := `
	SELECT EXISTS(
		SELECT 1
		FROM information_schema.tables
		WHERE UPPER(table_schema) = UPPER($1)
		AND UPPER(table_name) = UPPER($2));`
	items, err := jdb.Query(s.jdb, sql, schema, name)
	if err != nil {
		return false, err
	}

	if items.Count == 0 {
		return false, nil
	}

	return items.Bool(0, "exists"), nil
}

/**
* ddlTable
* @param model *jdb.Model
* @return string
**/
func (s *Params) ddlTable(model *jdb.Model) string {
	var columnsDef string
	for _, column := range model.Columns {
		if column.TypeColumn == jdb.TpColumn {
			def := strs.Format("\n\t%s %s DEFAULT %v", column.Name, s.typeData(column.TypeData), s.defaultValue(column.TypeData))
			columnsDef = strs.Append(columnsDef, def, ",")
		}
	}
	result := strs.Format("\nCREATE TABLE IF NOT EXISTS %s (%s\n);", tableName(model), columnsDef)

	return result
}

/**
* ddlTableRename
* @param oldName string
* @param newName string
* @return string
**/
func (s *Params) ddlTableRename(oldName, newName string) string {
	result := strs.Format(`ALTER TABLE %s RENAME TO %s;`, oldName, newName)

	return result
}

/**
* ddlTableInsertTo
* @param model *jdb.Model
* @param tableOrigin string
* @return string
**/
func (s *Params) ddlTableInsertTo(model *jdb.Model, tableOrigin string) string {
	fields := ""
	for _, column := range model.Columns {
		if column.TypeColumn == jdb.TpColumn {
			fields = strs.Append(fields, strs.Format("%s", column.Name), ", ")
		}
	}
	result := strs.Format("INSERT INTO %s (%s)\nSELECT %s FROM %s;", tableName(model), fields, fields, tableOrigin)

	return result
}

/**
* ddlTableDrop
* @param table string
* @return string
**/
func (s *Params) ddlTableDrop(table string) string {
	result := strs.Format("DROP TABLE IF EXISTS %s CASCADE;", table)

	return result
}

/**
* ddlTableEmpty
* @param table string
* @return string
**/
func (s *Params) ddlTableEmpty(table string) string {
	result := strs.Format("TRUNCATE TABLE %s CASCADE;", table)

	return result
}

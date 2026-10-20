package oracle

import (
	"slices"
	"sort"
	"strings"

	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/strs"
	"github.com/celsiainternet/elvis/utility"
	"github.com/celsiainternet/jdb/jdb"
)

/**
* jsonColumnTypes son los TypeData que se guardan como CLOB con CHECK (... IS JSON).
**/
var jsonColumnTypes = []jdb.TypeData{jdb.TypeDataObject, jdb.TypeDataMultiSelect, jdb.TypeDataGeometry}

/**
* clobColumnTypes son los TypeData que se guardan como CLOB (no admiten índice B-tree).
**/
var clobColumnTypes = []jdb.TypeData{jdb.TypeDataMemo, jdb.TypeDataObject, jdb.TypeDataMultiSelect, jdb.TypeDataGeometry, jdb.TypeDataFullText, jdb.TypeDataFilesMedia, jdb.TypeDataUrl, jdb.TypeDataAddress}

/**
* typeData
* @param tp jdb.TypeData
* @return string
**/
func (s *Params) typeData(tp jdb.TypeData) string {
	if slices.Contains(clobColumnTypes, tp) {
		return "CLOB"
	}

	switch tp {
	case jdb.TypeDataShortText, jdb.TypeDataKey, jdb.TypeDataState, jdb.TypeDataRelation, jdb.TypeDataRollup:
		return "VARCHAR2(80)"
	case jdb.TypeDataNumber:
		return "NUMBER(18,2)"
	case jdb.TypeDataInt:
		return "NUMBER(19)"
	case jdb.TypeDataPrecision:
		return "BINARY_DOUBLE"
	case jdb.TypeDataDateTime:
		return "TIMESTAMP"
	case jdb.TypeDataCheckbox:
		return "NUMBER(1)"
	case jdb.TypeDataBytes:
		return "BLOB"
	default:
		return "VARCHAR2(250)"
	}
}

/**
* defaultValue: nil cuando no hay DEFAULT; en Oracle '' es NULL, así que los textos no lo llevan.
* @param tp jdb.TypeData
* @return interface{}
**/
func (s *Params) defaultValue(tp jdb.TypeData) interface{} {
	switch tp {
	case jdb.TypeDataNumber, jdb.TypeDataPrecision, jdb.TypeDataInt:
		return 0
	case jdb.TypeDataDateTime:
		return "SYSTIMESTAMP"
	case jdb.TypeDataCheckbox:
		return quote(false)
	case jdb.TypeDataObject:
		return quote(et.Json{})
	case jdb.TypeDataMultiSelect:
		return quote([]et.Json{})
	case jdb.TypeDataGeometry:
		return quote(et.Json{
			"type":        "Point",
			"coordinates": []float64{0, 0},
		})
	case jdb.TypeDataState:
		return quote(utility.ACTIVE)
	default:
		return nil
	}
}

/**
* tableBase: el esquema de Oracle es el usuario; schema.table se aplana como schema_table
* dentro del usuario conectado.
* @param model *jdb.Model
* @return string
**/
func tableBase(model *jdb.Model) string {
	if model.Schema == "" {
		return model.Table
	}

	return strs.Format(`%s_%s`, model.Schema, model.Table)
}

/**
* tableName
* @param model *jdb.Model
* @return string
**/
func tableName(model *jdb.Model) string {
	return quoteIdent(tableBase(model))
}

/**
* objectName: nombres de índices y constraints, únicos por usuario; se prefijan con el esquema.
* @param model *jdb.Model, name string
* @return string
**/
func objectName(model *jdb.Model, name string) string {
	if model.Schema == "" {
		return quoteIdent(name)
	}

	return quoteIdent(strs.Format(`%s_%s`, model.Schema, name))
}

/**
* ddlTable: solo columnas; la llave primaria, las foráneas y los índices se agregan después
* (ddlTableIndex) para poder recrearlos al mutar la tabla.
* @param model *jdb.Model, table string
* @return string
**/
func (s *Params) ddlTable(model *jdb.Model, table string) string {
	defs := []string{}
	for _, column := range model.Columns {
		if column.TypeColumn != jdb.TpColumn {
			continue
		}

		def := strs.Format("\t%s %s", quoteIdent(column.Name), s.typeData(column.TypeData))
		if val := s.defaultValue(column.TypeData); val != nil {
			def = strs.Format("%s DEFAULT %v", def, val)
		}
		if slices.Contains(jsonColumnTypes, column.TypeData) {
			def = strs.Format("%s CHECK (%s IS JSON)", def, quoteIdent(column.Name))
		}
		defs = append(defs, def)
	}

	return strs.Format("CREATE TABLE %s (\n%s\n)", table, strings.Join(defs, ",\n"))
}

/**
* ddlPrimaryKey
* @param model *jdb.Model
* @return string
**/
func (s *Params) ddlPrimaryKey(model *jdb.Model) string {
	keys := []string{}
	for _, column := range model.PrimaryKeys {
		keys = append(keys, quoteIdent(column.Name))
	}

	if len(keys) == 0 {
		return ""
	}

	sort.Strings(keys)
	return strs.Format("ALTER TABLE %s ADD CONSTRAINT %s PRIMARY KEY (%s)", tableName(model), objectName(model, model.Name+"_pk"), strings.Join(keys, ", "))
}

/**
* ddlForeignKeys: Oracle no tiene ON UPDATE CASCADE; OnUpdateCascade se ignora.
* @param model *jdb.Model
* @return []string
**/
func (s *Params) ddlForeignKeys(model *jdb.Model) []string {
	names := []string{}
	for name := range model.ForeignKeys {
		names = append(names, name)
	}
	sort.Strings(names)

	result := []string{}
	for _, name := range names {
		relation := model.ForeignKeys[name]
		if relation.With == nil || len(relation.Fk) == 0 {
			continue
		}

		fks := []string{}
		for fk := range relation.Fk {
			fks = append(fks, fk)
		}
		sort.Strings(fks)

		keys := []string{}
		references := []string{}
		for _, fk := range fks {
			keys = append(keys, quoteIdent(fk))
			references = append(references, quoteIdent(relation.Fk[fk]))
		}

		def := strs.Format("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s(%s)", tableName(model), objectName(model, name), strings.Join(keys, ", "), tableName(relation.With), strings.Join(references, ", "))
		if relation.OnDeleteCascade {
			def += " ON DELETE CASCADE"
		}
		result = append(result, def)
	}

	return result
}

/**
* ddlTableRename
* @param table, newName string
* @return string
**/
func (s *Params) ddlTableRename(table, newName string) string {
	return strs.Format(`ALTER TABLE %s RENAME TO %s`, table, quoteIdent(newName))
}

/**
* ddlTableInsertTo: copia solo las columnas que existen en la tabla de origen.
* @param model *jdb.Model, target, origin string, current map[string]bool
* @return string
**/
func (s *Params) ddlTableInsertTo(model *jdb.Model, target, origin string, current map[string]bool) string {
	fields := []string{}
	for _, column := range model.Columns {
		if column.TypeColumn == jdb.TpColumn && current[column.Name] {
			fields = append(fields, quoteIdent(column.Name))
		}
	}

	if len(fields) == 0 {
		return ""
	}

	list := strings.Join(fields, ", ")
	return strs.Format("INSERT INTO %s (%s)\nSELECT %s FROM %s", target, list, list, origin)
}

/**
* ddlTableDrop
* @param table string
* @return string
**/
func (s *Params) ddlTableDrop(table string) string {
	return strs.Format("DROP TABLE %s CASCADE CONSTRAINTS PURGE", table)
}

/**
* ddlTableEmpty
* @param table string
* @return string
**/
func (s *Params) ddlTableEmpty(table string) string {
	return strs.Format("DELETE FROM %s", table)
}

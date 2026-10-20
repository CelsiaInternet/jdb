package sqlite

import (
	"sort"
	"strings"

	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/strs"
	"github.com/celsiainternet/elvis/utility"
	jdb "github.com/celsiainternet/jdb/jdb"
)

/**
* typeData
* @param tp jdb.TypeData
* @return string
**/
func (s *Params) typeData(tp jdb.TypeData) string {
	switch tp {
	case jdb.TypeDataNumber:
		return "NUMERIC"
	case jdb.TypeDataInt:
		return "INTEGER"
	case jdb.TypeDataPrecision:
		return "REAL"
	case jdb.TypeDataDateTime:
		return "DATETIME"
	case jdb.TypeDataCheckbox:
		return "BOOLEAN"
	case jdb.TypeDataBytes:
		return "BLOB"
	default:
		// Texto y JSON (Object, MultiSelect, Geometry, FullText) se guardan como TEXT
		return "TEXT"
	}
}

/**
* defaultValue
* @param tp jdb.TypeData
* @return interface{}
**/
func (s *Params) defaultValue(tp jdb.TypeData) interface{} {
	switch tp {
	case jdb.TypeDataNumber, jdb.TypeDataPrecision:
		return 0.0
	case jdb.TypeDataInt:
		return 0
	case jdb.TypeDataDateTime:
		return "CURRENT_TIMESTAMP"
	case jdb.TypeDataCheckbox:
		return quote(false)
	case jdb.TypeDataBytes:
		return "X''"
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
		return quote("")
	}
}

/**
* tableBase: SQLite no tiene esquemas, schema.table se aplana como schema_table.
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
* indexName: los nombres de índice son únicos en toda la base, se prefijan con el esquema.
* @param model *jdb.Model, name string
* @return string
**/
func indexName(model *jdb.Model, name string) string {
	if model.Schema == "" {
		return quoteIdent(name)
	}

	return quoteIdent(strs.Format(`%s_%s`, model.Schema, name))
}

/**
* ddlTable: SQLite no admite ALTER TABLE ADD CONSTRAINT, la llave primaria y las foráneas van en el CREATE.
* @param model *jdb.Model, table string
* @return string
**/
func (s *Params) ddlTable(model *jdb.Model, table string) string {
	defs := []string{}
	for _, column := range model.Columns {
		if column.TypeColumn == jdb.TpColumn {
			defs = append(defs, strs.Format("\t%s %s DEFAULT %v", quoteIdent(column.Name), s.typeData(column.TypeData), s.defaultValue(column.TypeData)))
		}
	}

	if pk := s.ddlPrimaryKey(model); pk != "" {
		defs = append(defs, pk)
	}
	defs = append(defs, s.ddlForeignKeys(model)...)

	return strs.Format("CREATE TABLE IF NOT EXISTS %s (\n%s\n);", table, strings.Join(defs, ",\n"))
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
	return strs.Format("\tPRIMARY KEY (%s)", strings.Join(keys, ", "))
}

/**
* ddlForeignKeys
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

		def := strs.Format("\tCONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s(%s)", quoteIdent(name), strings.Join(keys, ", "), tableName(relation.With), strings.Join(references, ", "))
		if relation.OnDeleteCascade {
			def += " ON DELETE CASCADE"
		}
		if relation.OnUpdateCascade {
			def += " ON UPDATE CASCADE"
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
	return strs.Format(`ALTER TABLE %s RENAME TO %s;`, table, quoteIdent(newName))
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
	return strs.Format("INSERT INTO %s (%s)\nSELECT %s FROM %s;", target, list, list, origin)
}

/**
* ddlTableDrop
* @param table string
* @return string
**/
func (s *Params) ddlTableDrop(table string) string {
	return strs.Format("DROP TABLE IF EXISTS %s;", table)
}

/**
* ddlTableEmpty: SQLite no tiene TRUNCATE.
* @param table string
* @return string
**/
func (s *Params) ddlTableEmpty(table string) string {
	return strs.Format("DELETE FROM %s;", table)
}

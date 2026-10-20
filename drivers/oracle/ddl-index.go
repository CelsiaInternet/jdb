package oracle

import (
	"slices"
	"sort"

	"github.com/celsiainternet/elvis/strs"
	"github.com/celsiainternet/jdb/jdb"
)

/**
* ddlIndex: los atributos se indexan por JSON_VALUE (índice basado en función); las columnas
* CLOB no admiten índice B-tree y se omiten.
* @param model *jdb.Model, name string, col *jdb.Column
* @return string
**/
func ddlIndex(model *jdb.Model, name string, col *jdb.Column) string {
	table := tableName(model)
	if col.TypeColumn == jdb.TpAtribute && model.SourceField != nil {
		return strs.Format(`CREATE INDEX %s ON %s(JSON_VALUE(%s, %s))`, objectName(model, name), table, quoteIdent(model.SourceField.Name), jsonPath(col.Name))
	}

	if slices.Contains(clobColumnTypes, col.TypeData) {
		return ""
	}

	return strs.Format(`CREATE INDEX %s ON %s(%s)`, objectName(model, name), table, quoteIdent(col.Name))
}

/**
* ddlUniqueIndex
* @param model *jdb.Model, name string, col *jdb.Column
* @return string
**/
func ddlUniqueIndex(model *jdb.Model, name string, col *jdb.Column) string {
	if col.TypeColumn != jdb.TpColumn || slices.Contains(clobColumnTypes, col.TypeData) {
		return ""
	}

	return strs.Format(`CREATE UNIQUE INDEX %s ON %s(%s)`, objectName(model, name), tableName(model), quoteIdent(col.Name))
}

/**
* ddlTableIndex: llave primaria, foráneas, índices y únicos, en ese orden.
* @param model *jdb.Model
* @return []string
**/
func (s *Params) ddlTableIndex(model *jdb.Model) []string {
	result := []string{}
	if pk := s.ddlPrimaryKey(model); pk != "" {
		result = append(result, pk)
	}
	result = append(result, s.ddlForeignKeys(model)...)

	names := []string{}
	for name := range model.Indices {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		index := model.Indices[name]
		if index.Column.TypeColumn == jdb.TpColumn || index.Column.TypeColumn == jdb.TpAtribute {
			if def := ddlIndex(model, name, index.Column); def != "" {
				result = append(result, def)
			}
		}
	}

	names = []string{}
	for name := range model.Uniques {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		if def := ddlUniqueIndex(model, name, model.Uniques[name]); def != "" {
			result = append(result, def)
		}
	}

	return result
}

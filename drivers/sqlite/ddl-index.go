package sqlite

import (
	"sort"

	"github.com/celsiainternet/elvis/strs"
	jdb "github.com/celsiainternet/jdb/jdb"
)

/**
* ddlIndex: los atributos se indexan por la misma expresión ->> que usa asField.
* @param model *jdb.Model, name string, col *jdb.Column
* @return string
**/
func ddlIndex(model *jdb.Model, name string, col *jdb.Column) string {
	table := tableName(model)
	if col.TypeColumn == jdb.TpAtribute && model.SourceField != nil {
		return strs.Format(`CREATE INDEX IF NOT EXISTS %s ON %s((%s ->> %s));`, indexName(model, name), table, quoteIdent(model.SourceField.Name), jsonPath(col.Name))
	}

	return strs.Format(`CREATE INDEX IF NOT EXISTS %s ON %s(%s);`, indexName(model, name), table, quoteIdent(col.Name))
}

/**
* ddlUniqueIndex
* @param model *jdb.Model, name string, col *jdb.Column
* @return string
**/
func ddlUniqueIndex(model *jdb.Model, name string, col *jdb.Column) string {
	if col.TypeColumn != jdb.TpColumn {
		return ""
	}

	return strs.Format(`CREATE UNIQUE INDEX IF NOT EXISTS %s ON %s(%s);`, indexName(model, name), tableName(model), quoteIdent(col.Name))
}

/**
* ddlTableIndex: índices y únicos; la llave primaria y las foráneas ya van en el CREATE TABLE.
* @param model *jdb.Model
* @return []string
**/
func (s *Params) ddlTableIndex(model *jdb.Model) []string {
	result := []string{}
	names := []string{}
	for name := range model.Indices {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		index := model.Indices[name]
		if index.Column.TypeColumn == jdb.TpColumn || index.Column.TypeColumn == jdb.TpAtribute {
			result = append(result, ddlIndex(model, name, index.Column))
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

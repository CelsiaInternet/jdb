package postgres

import (
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/celsiainternet/elvis/strs"
	"github.com/celsiainternet/jdb/jdb"
)

/* searchFunction: unaccent marcada IMMUTABLE (la de la extension no lo es) para poder indexarla. */
const searchFunction = "public.jdb_unaccent"

/* ddlSearchSetup: extensiones y funcion de la busqueda; idempotente. La funcion llama a unaccent con su esquema y
 * su diccionario calificados, para que el indice no dependa del search_path. */
const ddlSearchSetup = `
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS unaccent;
DO $jdb$
DECLARE
	ext_schema text;
BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_proc p JOIN pg_namespace n ON n.oid = p.pronamespace WHERE n.nspname = 'public' AND p.proname = 'jdb_unaccent') THEN
		SELECT n.nspname INTO ext_schema FROM pg_extension e JOIN pg_namespace n ON n.oid = e.extnamespace WHERE e.extname = 'unaccent';
		EXECUTE format('CREATE FUNCTION public.jdb_unaccent(text) RETURNS text LANGUAGE sql IMMUTABLE PARALLEL SAFE STRICT AS %L',
			format('SELECT %I.unaccent(%L::regdictionary, $1)', ext_schema, ext_schema || '.unaccent'));
	END IF;
END
$jdb$;`

/**
* searchPart: Un campo de la busqueda como texto, sin nulos. Igual en la consulta (con el alias de la tabla) y
* en el indice (sin alias): Postgres compara la expresion, no el texto.
* @param field jdb.Field
* @return string
**/
func searchPart(field jdb.Field) string {
	tableAlias := ""
	if field.Model != nil {
		tableAlias = field.Model.As
	}

	if len(field.Path) > 0 {
		column := strs.Append(tableAlias, field.Name, ".")
		path := strings.Join(field.Path, ",")
		return fmt.Sprintf(`COALESCE(%s#>>'{%s}', '')`, column, path)
	}

	if field.TypeColumn == jdb.TpColumn {
		column := strs.Append(tableAlias, field.Name, ".")
		return fmt.Sprintf(`COALESCE(%s::text, '')`, column)
	}

	if field.TypeColumn == jdb.TpAtribute && field.Model != nil && field.Model.Model != nil && field.Model.SourceField != nil {
		source := strs.Append(tableAlias, field.Model.SourceField.Name, ".")
		return fmt.Sprintf(`COALESCE(%s#>>'{%s}', '')`, source, field.Name)
	}

	return ""
}

/**
* searchExpr: Los campos de la busqueda unidos con un espacio, en minusculas y sin tildes; "" si no hay campos.
* Es la expresion del indice de trigramas (ddlSearch).
* @param fields []*jdb.Field
* @return string
**/
func searchExpr(fields []*jdb.Field) string {
	parts := []string{}
	for _, field := range fields {
		if field == nil {
			continue
		}

		part := searchPart(*field)
		if part != "" {
			parts = append(parts, part)
		}
	}

	if len(parts) == 0 {
		return ""
	}

	text := strings.Join(parts, " || ' ' || ")
	return fmt.Sprintf(`%s(lower(%s))`, searchFunction, text)
}

/**
* searchCondition: Condicion Search sobre un TextSearch: la expresion contiene la palabra, sin tildes ni
* mayusculas. Sin campos no encuentra nada.
* @param search jdb.TextSearch, val interface{}
* @return string
**/
func searchCondition(search jdb.TextSearch, val interface{}) string {
	expr := searchExpr(search.Fields)
	if expr == "" {
		return "FALSE"
	}

	return fmt.Sprintf(`%s LIKE '%%' || %s(lower(%v)) || '%%'`, expr, searchFunction, val)
}

/**
* searchIndexName: Nombre del indice de busqueda de la tabla, con un hash de la expresion: si cambian los campos
* se crea otro indice (el anterior queda y se borra a mano).
* @param model *jdb.Model, expr string
* @return string
**/
func searchIndexName(model *jdb.Model, expr string) string {
	hash := fnv.New32a()
	hash.Write([]byte(expr))
	sum := hash.Sum32()
	table := model.Table
	if len(table) > 40 {
		table = table[:40]
	}

	return fmt.Sprintf("%s_search_%08x", table, sum)
}

/**
* ddlSearch: Extensiones, funcion e indice GIN de trigramas de los campos de busqueda del modelo (DefineSearch);
* "" si no declaro campos. Todo es idempotente: se ejecuta en cada inicio, tambien con la tabla ya creada.
* @param model *jdb.Model
* @return string
**/
func ddlSearch(model *jdb.Model) string {
	fields := []*jdb.Field{}
	for _, name := range model.SearchFields {
		field := model.GetField(name)
		if field != nil {
			fields = append(fields, field)
		}
	}

	expr := searchExpr(fields)
	if expr == "" {
		return ""
	}

	name := searchIndexName(model, expr)
	table := tableName(model)
	index := fmt.Sprintf(`CREATE INDEX IF NOT EXISTS %s ON %s USING GIN ((%s) gin_trgm_ops);`, name, table, expr)
	return strs.Append(ddlSearchSetup, index, "\n")
}

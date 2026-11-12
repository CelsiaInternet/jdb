package postgres

import (
	"strings"
	"testing"

	"github.com/celsiainternet/jdb/jdb"
)

/**
* searchFields: nombre (columna) y correo (atributo en _data) con el alias de tabla dado.
* @param as string
* @return []*jdb.Field
**/
func searchFields(as string) []*jdb.Field {
	source := &jdb.Model{SourceField: &jdb.Column{Name: "_data"}}
	name := &jdb.Field{Model: &jdb.QlFrom{As: as, Model: source}, Name: "name", TypeColumn: jdb.TpColumn, TypeData: jdb.TypeDataText}
	email := &jdb.Field{Model: &jdb.QlFrom{As: as, Model: source}, Name: "email", TypeColumn: jdb.TpAtribute, TypeData: jdb.TypeDataText}
	return []*jdb.Field{name, email}
}

func TestWhereTextSearchSql(t *testing.T) {
	cases := []struct {
		name  string
		where func(where *jdb.QlWhere)
		want  string
	}{
		{"texto sin tildes ni mayusculas", func(where *jdb.QlWhere) {
			where.Where(&jdb.TextSearch{Fields: searchFields("A")}).Search("Gómez")
		}, `public.jdb_unaccent(lower(COALESCE(A.name::text, '') || ' ' || COALESCE(A._data#>>'{email}', ''))) LIKE '%' || public.jdb_unaccent(lower('Gómez')) || '%'`},
		{"sin campos no encuentra nada", func(where *jdb.QlWhere) {
			where.Where(&jdb.TextSearch{}).Search("ana")
		}, `FALSE`},
		{"search en un campo sigue siendo ILIKE", func(where *jdb.QlWhere) {
			where.Where(searchFields("A")[0]).Search("ana")
		}, `A.name ILIKE '%' || 'ana' || '%'`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			where := &jdb.QlWhere{}
			c.where(where)
			got := whereConditions(where)
			if normalize(got) != normalize(c.want) {
				t.Fatalf("got %s, want %s", got, c.want)
			}
		})
	}
}

func TestSearchIndexExpr(t *testing.T) {
	// El indice usa la misma expresion que la consulta, sin alias de tabla.
	query := searchExpr(searchFields("A"))
	index := searchExpr(searchFields(""))
	if strings.ReplaceAll(query, "A.", "") != index {
		t.Fatalf("query %s, index %s", query, index)
	}

	model := &jdb.Model{Table: "people"}
	name := searchIndexName(model, index)
	if !strings.HasPrefix(name, "people_search_") || len(name) != len("people_search_")+8 {
		t.Fatalf("got %s", name)
	}
	if other := searchIndexName(model, index+" "); other == name {
		t.Fatalf("misma expresion para campos distintos: %s", other)
	}
}

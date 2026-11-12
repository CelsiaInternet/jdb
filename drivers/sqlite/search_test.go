package sqlite

import (
	"testing"

	"github.com/celsiainternet/elvis/et"
	jdb "github.com/celsiainternet/jdb/jdb"
)

/**
* newSearchPeople: personas con nombre y apellido en columnas, correo en el source y busqueda sobre los tres.
* @param t *testing.T, db *jdb.DB
* @return *jdb.Model
**/
func newSearchPeople(t *testing.T, db *jdb.DB) *jdb.Model {
	schema := jdb.NewSchema(db, "test")
	model := jdb.NewModel(schema, "search_people", 1)
	model.DefineModel()
	model.DefineColumn("team", jdb.TypeDataText)
	model.DefineColumn("name", jdb.TypeDataText)
	model.DefineColumn("last_name", jdb.TypeDataText)
	model.DefineAtribute("email", jdb.TypeDataText)
	model.DefineSearch("name", "last_name", "email", "nope", jdb.CREATED_AT)
	if err := model.Init(); err != nil {
		t.Fatal(err)
	}

	rows := []et.Json{
		{"_id": "p1", "team": "t1", "name": "Ana", "last_name": "Gomez", "email": "ana@a.co"},
		{"_id": "p2", "team": "t1", "name": "Luis", "last_name": "Anaya", "email": "luis@a.co"},
		{"_id": "p3", "team": "t2", "name": "Ana", "last_name": "Perez", "email": "ana.p@b.co"},
		{"_id": "p4", "team": "t1", "name": "Marta", "email": "gomez.m@a.co"},
	}
	for _, row := range rows {
		if _, err := model.Insert(row).Exec(); err != nil {
			t.Fatal(err)
		}
	}

	return model
}

func TestDefineSearch(t *testing.T) {
	db := connect(t)
	people := newSearchPeople(t, db)

	want := []string{"name", "last_name", "email"}
	if len(people.SearchFields) != len(want) {
		t.Fatalf("got %v, want %v", people.SearchFields, want)
	}
	for i := range want {
		if people.SearchFields[i] != want[i] {
			t.Fatalf("got %v, want %v", people.SearchFields, want)
		}
	}
}

func TestSearchWords(t *testing.T) {
	got := jdb.SearchWords("  Ana  ana GOMEZ ")
	if len(got) != 2 || got[0] != "Ana" || got[1] != "GOMEZ" {
		t.Fatalf("got %v", got)
	}

	many := jdb.SearchWords("a b c d e f g h i j k l")
	if len(many) != jdb.SEARCH_MAX_WORDS {
		t.Fatalf("got %d words, want %d", len(many), jdb.SEARCH_MAX_WORDS)
	}
}

func TestWhereSearch(t *testing.T) {
	db := connect(t)
	people := newSearchPeople(t, db)

	cases := []struct {
		name string
		ql   func() *jdb.Ql
		want []string
	}{
		{"una palabra en columnas y atributo", func() *jdb.Ql {
			return people.WhereSearch("gomez")
		}, []string{"p1", "p4"}},
		{"todas las palabras, en cualquier campo", func() *jdb.Ql {
			return people.WhereSearch("ana gomez")
		}, []string{"p1"}},
		{"parte de una palabra", func() *jdb.Ql {
			return people.WhereSearch("aya")
		}, []string{"p2"}},
		{"respeta el AND con otra condicion", func() *jdb.Ql {
			return people.Where("team").Eq("t1").WhereSearch("ana")
		}, []string{"p1", "p2"}},
		{"texto vacio no filtra", func() *jdb.Ql {
			return people.WhereSearch("   ")
		}, []string{"p1", "p2", "p3", "p4"}},
		{"sin coincidencias", func() *jdb.Ql {
			return people.WhereSearch("zzz")
		}, []string{}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			items, err := c.ql().OrderBy(jdb.KEY).All()
			if err != nil {
				t.Fatal(err)
			}
			assertIds(t, ids(items), c.want)
		})
	}
}

func TestWhereSearchWithoutFields(t *testing.T) {
	db := connect(t)
	people := newPeople(t, db)

	items, err := people.WhereSearch("ana").All()
	if err != nil {
		t.Fatal(err)
	}
	assertIds(t, ids(items), []string{})
}

func TestRunQuerySearch(t *testing.T) {
	db := connect(t)
	people := newSearchPeople(t, db)

	cases := []struct {
		name  string
		query et.Json
		want  []string
	}{
		{"search en la raiz", et.Json{"search": "ana gomez"}, []string{"p1"}},
		{"search con wheres", et.Json{"search": "ana", "wheres": []et.Json{{"team": et.Json{"eq": "t2"}}}}, []string{"p3"}},
		{"operador search en un campo", et.Json{"wheres": []et.Json{{"email": et.Json{"search": "gomez"}}}}, []string{"p4"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c.query["order_by"] = et.Json{"asc": []string{jdb.KEY}}
			list, err := jdb.RunQuery(people, c.query)
			if err != nil {
				t.Fatal(err)
			}
			items := et.Items{Result: list.Result}
			assertIds(t, ids(items), c.want)
		})
	}
}

/**
* assertIds: Falla si got no es igual a want, en el mismo orden.
* @param t *testing.T, got, want []string
**/
func assertIds(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

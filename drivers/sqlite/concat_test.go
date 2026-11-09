package sqlite

import (
	"fmt"
	"testing"

	"github.com/celsiainternet/elvis/et"
	jdb "github.com/celsiainternet/jdb/jdb"
)

/**
* newPeople: personas con nombre y apellido en columnas y correo en el source (_data).
* @param t *testing.T, db *jdb.DB
* @return *jdb.Model
**/
func newPeople(t *testing.T, db *jdb.DB) *jdb.Model {
	schema := jdb.NewSchema(db, "test")
	model := jdb.NewModel(schema, "people", 1)
	model.DefineModel()
	model.DefineColumn("team", jdb.TypeDataText)
	model.DefineColumn("name", jdb.TypeDataText)
	model.DefineColumn("last_name", jdb.TypeDataText)
	model.DefineAtribute("email", jdb.TypeDataText)
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

func TestWhereConcat(t *testing.T) {
	db := connect(t)
	people := newPeople(t, db)

	cases := []struct {
		name string
		ql   func() *jdb.Ql
		want []string
	}{
		{"columnas y atributo", func() *jdb.Ql {
			return people.WhereConcat("name", "last_name", "email").Like("%gomez%")
		}, []string{"p1", "p4"}},
		{"cruza campos", func() *jdb.Ql {
			return people.WhereConcat("name", "last_name").Like("%ana gomez%")
		}, []string{"p1"}},
		{"respeta el AND con otra condicion", func() *jdb.Ql {
			return people.Where("team").Eq("t1").WhereConcat("name", "last_name").Like("%ana%")
		}, []string{"p1", "p2"}},
		{"nulo como texto vacio", func() *jdb.Ql {
			return people.WhereConcat("last_name", "name").Like("%marta%")
		}, []string{"p4"}},
		{"sin campos validos no encuentra nada", func() *jdb.Ql {
			return people.Where("team").Eq("t1").WhereConcat("nope").Like("%ana%")
		}, []string{}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			items, err := c.ql().OrderBy(jdb.KEY).All()
			if err != nil {
				t.Fatal(err)
			}
			got := ids(items)
			if len(got) != len(c.want) {
				t.Fatalf("got %v, want %v", got, c.want)
			}
			for i := range got {
				if got[i] != c.want[i] {
					t.Fatalf("got %v, want %v", got, c.want)
				}
			}
		})
	}
}

func TestRunQueryConcat(t *testing.T) {
	db := connect(t)
	people := newPeople(t, db)

	// El alcance (team t1) va primero: el concat no lo puede escapar
	result, err := jdb.RunQuery(people, et.Json{
		"wheres": []interface{}{
			et.Json{"team": et.Json{"eq": "t1"}},
			et.Json{"concat": et.Json{"fields": []interface{}{"name", "last_name", "email", "nope"}, "like": "%ana%"}},
		},
		"order_by": et.Json{"asc": []interface{}{jdb.KEY}},
	})
	if err != nil {
		t.Fatal(err)
	}
	got := []string{}
	for _, item := range result.Result {
		got = append(got, item.Str(jdb.KEY))
	}
	if len(got) != 2 || got[0] != "p1" || got[1] != "p2" || result.All != 2 {
		t.Fatalf("must return [p1 p2] of t1 with all=2, got %v all=%d", got, result.All)
	}
}

/**
* TestRunQueryGroupedConditions: concat dentro de {and} y de {or}, y una columna desconocida dentro de {and}, que se
* omite entera (antes su operador caia sobre la condicion anterior: team LIKE '%ana%').
* @param t *testing.T
**/
func TestRunQueryGroupedConditions(t *testing.T) {
	db := connect(t)
	people := newPeople(t, db)
	concat := et.Json{"fields": []interface{}{"name", "last_name"}, "like": "%ana%"}

	cases := []struct {
		name   string
		wheres []interface{}
		want   string
	}{
		{"concat en and", []interface{}{et.Json{"team": et.Json{"eq": "t1"}}, et.Json{"and": et.Json{"concat": concat}}}, "[p1 p2]"},
		{"concat en or", []interface{}{et.Json{"team": et.Json{"eq": "t2"}}, et.Json{"or": et.Json{"concat": et.Json{"fields": []interface{}{"name"}, "eq": "Marta"}}}}, "[p3 p4]"},
		{"columna desconocida en and", []interface{}{et.Json{"team": et.Json{"eq": "t2"}}, et.Json{"and": et.Json{"nope": et.Json{"like": "%x%"}}}}, "[p3]"},
	}
	for _, c := range cases {
		result, err := jdb.RunQuery(people, et.Json{"wheres": c.wheres, "order_by": et.Json{"asc": []interface{}{jdb.KEY}}})
		if err != nil {
			t.Fatal(err)
		}
		got := []string{}
		for _, item := range result.Result {
			got = append(got, item.Str(jdb.KEY))
		}
		if fmt.Sprint(got) != c.want {
			t.Errorf("%s: got %v, want %s", c.name, got, c.want)
		}
	}
}

package sqlite

import (
	"testing"

	"github.com/celsiainternet/elvis/et"
	jdb "github.com/celsiainternet/jdb/jdb"
)

/**
* newSchedules: agendamientos con source (_data) y una columna JSON (meta) para filtrar por rutas JSON.
* @param t *testing.T, db *jdb.DB
* @return *jdb.Model
**/
func newSchedules(t *testing.T, db *jdb.DB) *jdb.Model {
	schema := jdb.NewSchema(db, "test")
	model := jdb.NewModel(schema, "schedules", 1)
	model.DefineModel()
	model.DefineColumn("customer_id", jdb.TypeDataText)
	model.DefineColumn("start_date", jdb.TypeDataText)
	model.DefineColumn("meta", jdb.TypeDataObject)
	if err := model.Init(); err != nil {
		t.Fatal(err)
	}

	rows := []et.Json{
		{"_id": "s1", "customer_id": "c1", "start_date": "2026-01-01", "meta": et.Json{"level": "alto"}, "sync_external": et.Json{"oss_order_id": "O-1", "attempts": 1}},
		{"_id": "s2", "customer_id": "c1", "start_date": "2026-02-01", "meta": et.Json{"level": "bajo"}, "sync_external": et.Json{"oss_order_id": "O-2", "attempts": 3}},
		{"_id": "s3", "customer_id": "c1", "start_date": "2026-03-01", "sync_external": et.Json{"oss_order_id": "O-2", "attempts": 5}},
		{"_id": "s4", "customer_id": "c2", "start_date": "2026-04-01", "sync_external": et.Json{"oss_order_id": "O-2", "attempts": 2}},
	}
	for _, row := range rows {
		if _, err := model.Insert(row).Exec(); err != nil {
			t.Fatal(err)
		}
	}

	return model
}

/**
* ids: _id de las filas, en orden.
* @param items et.Items
* @return []string
**/
func ids(items et.Items) []string {
	result := []string{}
	for _, item := range items.Result {
		result = append(result, item.Str(jdb.KEY))
	}
	return result
}

func TestWhereJsonPath(t *testing.T) {
	db := connect(t)
	schedules := newSchedules(t, db)

	cases := []struct {
		name  string
		field string
	}{
		{"flechas de Postgres", "_data->'sync_external'->>'oss_order_id'"},
		{"puntos desde la columna", "_data.sync_external.oss_order_id"},
		{"puntos dentro del source", "sync_external.oss_order_id"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			items, err := schedules.
				Where("customer_id").Eq("c1").
				And(c.field).Eq("O-2").
				OrderByDesc("start_date").
				All()
			if err != nil {
				t.Fatal(err)
			}

			got := ids(items)
			if len(got) != 2 || got[0] != "s3" || got[1] != "s2" {
				t.Fatalf("must return [s3 s2], got %v", got)
			}
		})
	}
}

func TestWhereJsonPathTypes(t *testing.T) {
	db := connect(t)
	schedules := newSchedules(t, db)

	// Numero: compara como numero, no como texto.
	items, err := schedules.Where("sync_external.attempts").More(2).OrderBy(jdb.KEY).All()
	if err != nil {
		t.Fatal(err)
	}
	if got := ids(items); len(got) != 2 || got[0] != "s2" || got[1] != "s3" {
		t.Errorf("attempts > 2: must return [s2 s3], got %v", got)
	}

	// Ruta dentro de una columna JSON que no es el source.
	items, err = schedules.Where("meta.level").Eq("alto").All()
	if err != nil {
		t.Fatal(err)
	}
	if got := ids(items); len(got) != 1 || got[0] != "s1" {
		t.Errorf("meta.level: must return [s1], got %v", got)
	}

	// In sobre una ruta.
	items, err = schedules.Where("sync_external.oss_order_id").In("O-1", "O-9").All()
	if err != nil {
		t.Fatal(err)
	}
	if got := ids(items); len(got) != 1 || got[0] != "s1" {
		t.Errorf("in: must return [s1], got %v", got)
	}
}

func TestJsonPathRejectsUnsafeKeys(t *testing.T) {
	db := connect(t)
	schedules := newSchedules(t, db)

	// Una llave con comillas o una ruta sobre una columna que no es JSON no se resuelven como campo:
	// quedan como antes (literal), asi que no filtran nada y no inyectan SQL.
	for _, name := range []string{"_data.sync_external.oss'order", "customer_id.x", "_data->'a''b'"} {
		items, err := schedules.Where(name).Eq("O-2").All()
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if items.Count != 0 {
			t.Errorf("%s: must not match rows, got %v", name, ids(items))
		}
	}

	all, err := schedules.Where(jdb.KEY).NotNull().All()
	if err != nil {
		t.Fatal(err)
	}
	if all.Count != 4 {
		t.Errorf("the table must keep its 4 rows, got %d", all.Count)
	}
}

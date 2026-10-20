package postgres

import (
	"strings"
	"testing"

	"github.com/celsiainternet/elvis/et"
	jdb "github.com/celsiainternet/jdb/jdb"
)

/**
* newTestModel: Crea un modelo en memoria, sin conexión, con source y una columna name.
* @param name string
* @param integrity bool
* @return *jdb.Model
**/
func newTestModel(name string, integrity bool) *jdb.Model {
	db := &jdb.DB{Name: "test"}
	schema := jdb.NewSchema(db, "test")
	model := jdb.NewModel(schema, name, 1)
	model.DefineProjectModel()
	model.DefineColumn("name", jdb.TypeDataText)
	if integrity {
		model.DefineIntegrity()
	}

	return model
}

/**
* sourceArg: Devuelve el argumento jsonb del source como et.Json.
* @param t *testing.T
* @param args []any
* @return et.Json
**/
func sourceArg(t *testing.T, args []any) et.Json {
	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %d: %v", len(args), args)
	}

	str, ok := args[0].(string)
	if !ok {
		t.Fatalf("expected string arg, got %T", args[0])
	}

	result, err := et.Object(str)
	if err != nil {
		t.Fatal(err)
	}

	return result
}

func TestSqlInsertDynamicAtributes(t *testing.T) {
	model := newTestModel("insert_dynamic", false)
	data := et.Json{"_id": "1", "name": "Ana", "nickname": "anita", "age": 30}
	command := jdb.NewCommand(model, []et.Json{data}, jdb.Insert)
	command.New = data

	sql, args := (&Params{}).sqlInsert(command)
	source := sourceArg(t, args)
	if source.Str("nickname") != "anita" {
		t.Errorf("nickname not in source: %v\n%s", source, sql)
	}

	if source.Int("age") != 30 {
		t.Errorf("age not in source: %v\n%s", source, sql)
	}

	if source["name"] != nil {
		t.Errorf("column name must not be in source: %v", source)
	}
}

func TestSqlUpdateDynamicAtributes(t *testing.T) {
	model := newTestModel("update_dynamic", false)
	data := et.Json{"name": "Ana", "nickname": "anita", "age": 30}
	command := jdb.NewCommand(model, []et.Json{data}, jdb.Update)
	command.New = data
	command.Where("_id").Eq("1")

	sql, args := (&Params{}).sqlUpdate(command)
	if !strings.Contains(sql, "_data = COALESCE(_data, '{}'::jsonb) || $1::jsonb") {
		t.Errorf("source merge missing:\n%s", sql)
	}

	source := sourceArg(t, args)
	if source.Str("nickname") != "anita" {
		t.Errorf("nickname not in source: %v", source)
	}

	// JSON decodifica los números como float64; lo que importa es que no llegue como string
	if _, ok := source["age"].(float64); !ok {
		t.Errorf("age must keep its numeric type, got %T", source["age"])
	}
}

func TestSqlInsertIntegrityDropsAtributes(t *testing.T) {
	model := newTestModel("insert_integrity", true)
	data := et.Json{"_id": "1", "name": "Ana", "nickname": "anita"}
	command := jdb.NewCommand(model, []et.Json{data}, jdb.Insert)
	command.New = data

	_, args := (&Params{}).sqlInsert(command)
	source := sourceArg(t, args)
	if len(source) != 0 {
		t.Errorf("integrity model must not store extra keys: %v", source)
	}
}

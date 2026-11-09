package sqlite

import (
	"testing"

	"github.com/celsiainternet/elvis/et"
	jdb "github.com/celsiainternet/jdb/jdb"
)

/**
* newAccounts: cuentas con una columna oculta (secret) y otra visible (name).
* @param t *testing.T, db *jdb.DB
* @return *jdb.Model
**/
func newAccounts(t *testing.T, db *jdb.DB) *jdb.Model {
	schema := jdb.NewSchema(db, "test")
	model := jdb.NewModel(schema, "accounts", 1)
	model.DefineModel()
	model.DefineColumn("name", jdb.TypeDataText)
	model.DefineColumn("secret", jdb.TypeDataText)
	model.DefineHidden("secret")
	if err := model.Init(); err != nil {
		t.Fatal(err)
	}
	if _, err := model.Insert(et.Json{jdb.KEY: "a1", "name": "Ana", "secret": "s3cr3t"}).Exec(); err != nil {
		t.Fatal(err)
	}
	return model
}

func TestHiddenOnlyWhenSelected(t *testing.T) {
	db := connect(t)
	accounts := newAccounts(t, db)

	// Sin select: la oculta no sale
	item, err := accounts.Where(jdb.KEY).Eq("a1").One()
	if err != nil {
		t.Fatal(err)
	}
	if item.Result.Str("name") != "Ana" || item.Result["secret"] != nil {
		t.Errorf("sin select: %v", item.Result)
	}

	// Pedida por nombre: sale
	item, err = jdb.From(accounts).Data("secret").Where(jdb.KEY).Eq("a1").One()
	if err != nil {
		t.Fatal(err)
	}
	if item.Result.Str("secret") != "s3cr3t" {
		t.Errorf("con Data(secret): %v", item.Result)
	}

	// jsql sin selects: no sale; con selects: sale (el que expone jsql decide si lo permite)
	list, err := jdb.RunQuery(accounts, et.Json{})
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Result) != 1 || list.Result[0]["secret"] != nil {
		t.Errorf("RunQuery sin selects: %v", list.Result)
	}

	// Filtrar por la oculta sigue funcionando (el login compara el hash con WHERE)
	exists, err := accounts.Where("secret").Eq("s3cr3t").ItExists()
	if err != nil || !exists {
		t.Errorf("where por oculta: %v %v", exists, err)
	}

	// Ql.Hidden oculta una columna mas en esa consulta
	item, err = jdb.From(accounts).Hidden("name").Where(jdb.KEY).Eq("a1").One()
	if err != nil || item.Result["name"] != nil {
		t.Errorf("Ql.Hidden: %v %v", item.Result, err)
	}
}

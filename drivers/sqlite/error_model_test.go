package sqlite

import (
	"errors"
	"strings"
	"testing"

	"github.com/celsiainternet/elvis/et"
	jdb "github.com/celsiainternet/jdb/jdb"
)

/**
* newBroken: modelo cuya columna name se borra de la tabla despues de Init, como una columna nueva
* que no existe en una base vieja (jdb solo hace CREATE TABLE IF NOT EXISTS).
* @param t *testing.T, db *jdb.DB
* @return *jdb.Model
**/
func newBroken(t *testing.T, db *jdb.DB) *jdb.Model {
	schema := jdb.NewSchema(db, "test")
	model := jdb.NewModel(schema, "broken", 1)
	model.DefineModel()
	model.DefineColumn("name", jdb.TypeDataText)
	if err := model.Init(); err != nil {
		t.Fatal(err)
	}
	if _, err := model.Insert(et.Json{jdb.KEY: "a", "name": "uno"}).Exec(); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Query(`ALTER TABLE "test_broken" DROP COLUMN "name";`); err != nil {
		t.Fatal(err)
	}

	return model
}

/* hasModel: el error empieza con el modelo y conserva el texto del driver */
func hasModel(t *testing.T, op string, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: se esperaba un error", op)
	}
	if !strings.HasPrefix(err.Error(), "test.broken") {
		t.Errorf("%s: el error no dice el modelo: %v", op, err)
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("%s: el error perdio el texto del driver: %v", op, err)
	}
	if strings.Count(err.Error(), "test.broken") != 1 {
		t.Errorf("%s: el modelo se repite: %v", op, err)
	}
}

func TestErrorsNameTheModel(t *testing.T) {
	db := connect(t)
	model := newBroken(t, db)

	_, err := model.Where("name").Eq("uno").All()
	hasModel(t, "All", err)

	_, err = model.Where("name").Eq("uno").ItExists()
	hasModel(t, "ItExists", err)

	_, err = model.Where("name").Eq("uno").Counted()
	hasModel(t, "Counted", err)

	_, err = model.Insert(et.Json{jdb.KEY: "b", "name": "dos"}).Exec()
	hasModel(t, "Insert", err)

	_, err = model.Update(et.Json{"name": "tres"}).Where(jdb.KEY).Eq("a").Exec()
	hasModel(t, "Update", err)
}

func TestHookErrorsKeepTheirMessage(t *testing.T) {
	db := connect(t)
	model := newTxItems(t, db)
	fail := errors.New("el nombre ya existe")
	model.BeforeInsert(func(tx *jdb.Tx, data et.Json) error {
		return fail
	})

	_, err := model.Insert(et.Json{jdb.KEY: "a", "name": "uno"}).Exec()
	if !errors.Is(err, fail) || err.Error() != fail.Error() {
		t.Fatalf("Insert = %v, want %v sin prefijo", err, fail)
	}
}

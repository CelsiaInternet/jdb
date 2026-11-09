package sqlite

import (
	"errors"
	"testing"

	"github.com/celsiainternet/elvis/et"
	jdb "github.com/celsiainternet/jdb/jdb"
)

/**
* newTxItems: modelo simple para las pruebas de transacciones.
* @param t *testing.T, db *jdb.DB
* @return *jdb.Model
**/
func newTxItems(t *testing.T, db *jdb.DB) *jdb.Model {
	schema := jdb.NewSchema(db, "test")
	model := jdb.NewModel(schema, "tx_items", 1)
	model.DefineModel()
	model.DefineColumn("name", jdb.TypeDataText)
	if err := model.Init(); err != nil {
		t.Fatal(err)
	}

	return model
}

func countTx(t *testing.T, model *jdb.Model) int {
	t.Helper()
	n, err := model.Counted()
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func TestTxBeginEndCommit(t *testing.T) {
	db := connect(t)
	model := newTxItems(t, db)

	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"a", "b"} {
		if _, err := model.Insert(et.Json{"_id": id, "name": id}).ExecTx(tx); err != nil {
			t.Fatal(err)
		}
	}
	if err := tx.End(nil); err != nil {
		t.Fatal(err)
	}

	if n := countTx(t, model); n != 2 {
		t.Errorf("filas tras End(nil): %d, want 2", n)
	}
}

func TestTxBeginEndRollback(t *testing.T) {
	db := connect(t)
	model := newTxItems(t, db)

	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := model.Insert(et.Json{"_id": "a", "name": "a"}).ExecTx(tx); err != nil {
		t.Fatal(err)
	}
	fail := errors.New("fallo de negocio")
	if err := tx.End(fail); !errors.Is(err, fail) {
		t.Errorf("End(err) = %v, want %v", err, fail)
	}

	if n := countTx(t, model); n != 0 {
		t.Errorf("filas tras End(err): %d, want 0", n)
	}
}

func TestBeginWithoutConnection(t *testing.T) {
	var db *jdb.DB
	if _, err := db.Begin(); err == nil {
		t.Error("Begin sin conexion deberia fallar")
	}
}

/*
Un error en un update con transaccion propia la revierte; antes quedaba en un err sombreado y

	se confirmaba igual
*/
func TestExecTxRollbackOnUpdateError(t *testing.T) {
	db := connect(t)
	model := newTxItems(t, db)
	if _, err := model.Insert(et.Json{"_id": "a", "name": "original"}).Exec(); err != nil {
		t.Fatal(err)
	}

	fail := errors.New("after update fallo")
	model.AfterUpdate(func(tx *jdb.Tx, data et.Json) error {
		return fail
	})

	_, err := model.
		Update(et.Json{"name": "cambiado"}).
		Where(jdb.KEY).Eq("a").
		Exec()
	if !errors.Is(err, fail) {
		t.Fatalf("Exec = %v, want %v", err, fail)
	}

	item, err := model.Where(jdb.KEY).Eq("a").One()
	if err != nil {
		t.Fatal(err)
	}
	if item.Str("name") != "original" {
		t.Errorf("name = %q: el update con error se confirmo", item.Str("name"))
	}
}

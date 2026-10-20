package sqlite

import (
	"strings"
	"testing"

	"github.com/celsiainternet/elvis/et"
	jdb "github.com/celsiainternet/jdb/jdb"
)

/**
* connect: base SQLite en memoria, una por test.
* @param t *testing.T
* @return *jdb.DB
**/
func connect(t *testing.T) *jdb.DB {
	db, err := jdb.ConnectTo(jdb.ConnectParams{
		Driver: jdb.SqliteDriver,
		Name:   "sqlite_" + t.Name(),
		Params: &Connection{Database: ":memory:"},
	})
	if err != nil {
		t.Fatal(err)
	}

	return db
}

/**
* newUsers: modelo con columnas, un atributo declarado y un booleano.
* @param t *testing.T, db *jdb.DB, integrity bool
* @return *jdb.Model
**/
func newUsers(t *testing.T, db *jdb.DB, integrity bool) *jdb.Model {
	schema := jdb.NewSchema(db, "test")
	model := jdb.NewModel(schema, "users", 1)
	model.DefineModel()
	model.DefineColumn("name", jdb.TypeDataText)
	model.DefineColumn("age", jdb.TypeDataInt)
	model.DefineColumn("active", jdb.TypeDataCheckbox)
	model.DefineColumn("tags", jdb.TypeDataObject)
	model.DefineAtribute("city", jdb.TypeDataText)
	model.DefineIndex(true, "name", "city")
	if integrity {
		model.DefineIntegrity()
	}

	err := model.Init()
	if err != nil {
		t.Fatal(err)
	}

	return model
}

func TestCrudDynamicAtributes(t *testing.T) {
	db := connect(t)
	users := newUsers(t, db, false)

	items, err := users.Insert(et.Json{
		"_id":      "u1",
		"name":     "Ana",
		"age":      30,
		"active":   true,
		"tags":     et.Json{"role": "admin"},
		"city":     "Bogotá",
		"nickname": "anita",
		"score":    7.5,
	}).Exec()
	if err != nil {
		t.Fatal(err)
	}

	if items.Count != 1 {
		t.Fatalf("insert must return 1 row, got %d", items.Count)
	}

	item := items.Result[0]
	if item.Str("nickname") != "anita" || item.Str("city") != "Bogotá" || item.Str("name") != "Ana" {
		t.Errorf("insert result: %v", item)
	}

	if item["active"] != true {
		t.Errorf("active must be a bool, got %T %v", item["active"], item["active"])
	}

	if item.Json("tags").Str("role") != "admin" {
		t.Errorf("tags must be an object, got %T %v", item["tags"], item["tags"])
	}

	items, err = users.Update(et.Json{"nickname": "ani", "age": 31, "level": 3}).Where("_id").Eq("u1").Exec()
	if err != nil {
		t.Fatal(err)
	}

	if items.Count != 1 {
		t.Fatalf("update must return 1 row, got %d", items.Count)
	}

	item = items.Result[0]
	if item.Str("nickname") != "ani" || item.Int("age") != 31 || item.Int("level") != 3 || item.Str("city") != "Bogotá" {
		t.Errorf("update result: %v", item)
	}

	if _, ok := item["level"].(float64); !ok {
		t.Errorf("level must keep its numeric type, got %T", item["level"])
	}

	found, err := users.Data().Where("city").Eq("Bogotá").All()
	if err != nil {
		t.Fatal(err)
	}

	if found.Count != 1 || found.Result[0].Str("nickname") != "ani" {
		t.Errorf("select by atribute: %v", found)
	}

	rows, err := users.Select("name", "active").Where("name").Like("an%").All()
	if err != nil {
		t.Fatal(err)
	}

	if rows.Count != 1 || rows.Result[0]["active"] != true {
		t.Errorf("select columns: %v", rows)
	}

	count, err := users.Where("age").More(18).Counted()
	if err != nil {
		t.Fatal(err)
	}

	if count != 1 {
		t.Errorf("count: %d", count)
	}

	items, err = users.Delete("_id").Eq("u1").Exec()
	if err != nil {
		t.Fatal(err)
	}

	if items.Count != 1 || items.Result[0].Str("nickname") != "ani" {
		t.Errorf("delete result: %v", items)
	}

	count, err = users.Where("_id").Eq("u1").Counted()
	if err != nil {
		t.Fatal(err)
	}

	if count != 0 {
		t.Errorf("row not deleted")
	}
}

func TestInsertIntegrityDropsAtributes(t *testing.T) {
	db := connect(t)
	users := newUsers(t, db, true)

	items, err := users.Insert(et.Json{"_id": "u1", "name": "Ana", "city": "Cali", "nickname": "anita"}).Exec()
	if err != nil {
		t.Fatal(err)
	}

	item := items.Result[0]
	if item["nickname"] != nil {
		t.Errorf("integrity model must not store extra keys: %v", item)
	}

	if item.Str("city") != "Cali" {
		t.Errorf("declared atribute must be stored: %v", item)
	}
}

func TestUpdateRequiresWhere(t *testing.T) {
	db := connect(t)
	users := newUsers(t, db, false)
	command := jdb.NewCommand(users, []et.Json{{"name": "x"}}, jdb.Update)
	command.New = et.Json{"name": "x"}

	_, err := db.Command(command)
	if err == nil {
		t.Fatal("update without where must fail")
	}
}

func TestMutateModelKeepsData(t *testing.T) {
	db := connect(t)
	users := newUsers(t, db, false)

	_, err := users.Insert(et.Json{"_id": "u1", "name": "Ana", "nickname": "anita"}).Exec()
	if err != nil {
		t.Fatal(err)
	}

	users.DefineColumn("email", jdb.TypeDataEmail)
	err = db.MutateModel(users)
	if err != nil {
		t.Fatal(err)
	}

	item, err := users.Data().Where("_id").Eq("u1").One()
	if err != nil {
		t.Fatal(err)
	}

	if item.Str("name") != "Ana" || item.Str("nickname") != "anita" {
		t.Errorf("data lost after mutate: %v", item)
	}

	if _, ok := item.Result["email"]; !ok {
		t.Errorf("new column missing after mutate: %v", item)
	}
}

func TestSqlUpdateBindsAtributes(t *testing.T) {
	db := connect(t)
	users := newUsers(t, db, false)
	command := jdb.NewCommand(users, []et.Json{}, jdb.Update)
	command.New = et.Json{"nick'name": "x"}
	command.Where("_id").Eq("1")

	sql, args := (&Params{}).sqlUpdate(command)
	if strings.Contains(sql, "nick'name") {
		t.Errorf("atribute key must be bound, not inlined:\n%s", sql)
	}

	if !strings.Contains(sql, `UPDATE "test_users" AS A`) {
		t.Errorf("update must alias the table so WHERE A.x resolves:\n%s", sql)
	}

	if len(args) != 2 || args[0] != `$."nick'name"` || args[1] != `"x"` {
		t.Errorf("args: %v", args)
	}
}

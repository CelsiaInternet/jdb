package sqlite

import (
	"encoding/json"
	"testing"

	"github.com/celsiainternet/elvis/et"
	jdb "github.com/celsiainternet/jdb/jdb"
)

/**
* newRollupModels: equipos y roles; roles trae del equipo su nombre (rollup de un campo) y nombre y logo (objeto).
* @param t *testing.T, db *jdb.DB
* @return *jdb.Model
**/
func newRollupModels(t *testing.T, db *jdb.DB) *jdb.Model {
	schema := jdb.NewSchema(db, "test")
	teams := jdb.NewModel(schema, "teams", 1)
	teams.DefineModel()
	teams.DefineColumn("name", jdb.TypeDataText)
	teams.DefineColumn("logo", jdb.TypeDataText)
	if err := teams.Init(); err != nil {
		t.Fatal(err)
	}

	roles := jdb.NewModel(schema, "roles", 1)
	roles.DefineModel()
	roles.DefineColumn("user_id", jdb.TypeDataText)
	roles.DefineColumn("team_id", jdb.TypeDataText)
	roles.DefineRollup("team_name", teams.Name, map[string]string{"team_id": jdb.KEY}, []string{"name"})
	roles.DefineObject("team", teams.Name, map[string]string{"team_id": jdb.KEY}, []string{"name", "logo"})
	if err := roles.Init(); err != nil {
		t.Fatal(err)
	}

	for _, team := range []et.Json{{"_id": "t1", "name": "Ventas", "logo": "v.png"}, {"_id": "t2", "name": "Soporte", "logo": "s.png"}} {
		if _, err := teams.Insert(team).Exec(); err != nil {
			t.Fatal(err)
		}
	}
	for _, role := range []et.Json{{"_id": "r1", "user_id": "u1", "team_id": "t1"}, {"_id": "r2", "user_id": "u2", "team_id": "t1"}, {"_id": "r3", "user_id": "u3", "team_id": "t9"}, {"_id": "r4", "user_id": "u4"}} {
		if _, err := roles.Insert(role).Exec(); err != nil {
			t.Fatal(err)
		}
	}

	return roles
}

func TestRollupResolvesOnSelect(t *testing.T) {
	db := connect(t)
	roles := newRollupModels(t, db)

	items, err := roles.Select(jdb.KEY, "team_id", "team_name", "team").OrderBy(jdb.KEY).All()
	if err != nil {
		t.Fatal(err)
	}

	if items.Count != 4 {
		t.Fatalf("must return 4 roles, got %d", items.Count)
	}

	byId := map[string]et.Json{}
	for _, item := range items.Result {
		byId[item.Str(jdb.KEY)] = item
	}

	// Un campo con ShowAtrib: el valor queda en la columna del rollup.
	if byId["r1"].Str("team_name") != "Ventas" || byId["r2"].Str("team_name") != "Ventas" {
		t.Errorf("team_name: r1=%v r2=%v", byId["r1"]["team_name"], byId["r2"]["team_name"])
	}

	// DefineObject: objeto con los campos.
	team := byId["r1"].Json("team")
	if team.Str("name") != "Ventas" || team.Str("logo") != "v.png" {
		t.Errorf("team object: %v", byId["r1"]["team"])
	}

	// Las filas con la misma llave comparten la consulta pero no el objeto.
	byId["r1"].Json("team")["name"] = "cambiado"
	if byId["r2"].Json("team").Str("name") != "Ventas" {
		t.Errorf("rows must not share the rollup object: %v", byId["r2"]["team"])
	}

	// Sin registro relacionado o sin llave: campos en nil.
	if byId["r3"]["team_name"] != nil || byId["r4"]["team_name"] != nil {
		t.Errorf("missing team must be nil: r3=%v r4=%v", byId["r3"]["team_name"], byId["r4"]["team_name"])
	}
	if team4 := byId["r4"].Json("team"); team4["name"] != nil || team4["logo"] != nil {
		t.Errorf("missing team object must have nil fields: %v", byId["r4"]["team"])
	}
}

func TestRollupNotRequestedIsNotResolved(t *testing.T) {
	db := connect(t)
	roles := newRollupModels(t, db)

	item, err := roles.Where(jdb.KEY).Eq("r1").One()
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := item.Result["team_name"]; ok {
		t.Errorf("rollup not requested must not be resolved: %v", item.Result)
	}
}

func TestRollupSerialize(t *testing.T) {
	db := connect(t)
	roles := newRollupModels(t, db)

	definition, err := json.Marshal(roles)
	if err != nil {
		t.Fatal(err)
	}

	loaded := struct {
		Rollup map[string]*jdb.Rollup `json:"rollup"`
	}{}
	if err := json.Unmarshal(definition, &loaded); err != nil {
		t.Fatal(err)
	}

	team := loaded.Rollup["team"]
	if team == nil || team.Name != "team" || team.WithName != "teams" || team.Show != jdb.ShowObject || team.Fk["team_id"] != jdb.KEY {
		t.Fatalf("serialized rollup: %s", definition)
	}

	if loaded.Rollup["team_name"].Show != jdb.ShowAtrib {
		t.Errorf("team_name show: %v", loaded.Rollup["team_name"].Show)
	}

	// Una definicion vieja con show numerico se sigue leyendo.
	old := jdb.Rollup{}
	if err := json.Unmarshal([]byte(`{"show":1}`), &old); err != nil || old.Show != jdb.ShowObject {
		t.Errorf("numeric show: %v %v", old.Show, err)
	}

	describe := roles.Describe()
	if _, ok := describe["rollup"]; !ok {
		t.Errorf("describe must include rollup: %v", describe)
	}
}

/**
* newKeyModels: tipos de identificacion con varios registros por codigo (key) y personas que los referencian.
* @param t *testing.T, db *jdb.DB
* @return *jdb.Model, *jdb.Model
**/
func newKeyModels(t *testing.T, db *jdb.DB) (*jdb.Model, *jdb.Model) {
	schema := jdb.NewSchema(db, "test")
	types := jdb.NewModel(schema, "id_types", 1)
	types.DefineModel()
	types.DefineColumn("code", jdb.TypeDataText)
	types.DefineColumn("key", jdb.TypeDataText)
	types.DefineColumn("name", jdb.TypeDataText)
	if err := types.Init(); err != nil {
		t.Fatal(err)
	}

	people := jdb.NewModel(schema, "people", 1)
	people.DefineModel()
	people.DefineColumn("type_code", jdb.TypeDataText)
	if err := people.Init(); err != nil {
		t.Fatal(err)
	}

	for _, row := range []et.Json{
		{"_id": "k1", "code": "1", "key": "CC", "name": "Cedula"},
		{"_id": "k2", "code": "1", "key": "TI", "name": "Tarjeta"},
		{"_id": "k3", "code": "2", "key": "NIT", "name": "Nit"},
	} {
		if _, err := types.Insert(row).Exec(); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := people.Insert(et.Json{"_id": "p1", "type_code": "1"}).Exec(); err != nil {
		t.Fatal(err)
	}

	return types, people
}

func TestRollupWheres(t *testing.T) {
	db := connect(t)
	types, people := newKeyModels(t, db)

	fk := map[string]string{"type_code": "code"}
	if _, err := people.DefineRollup("type_cc", types.Name, fk, []string{"name"}).SetWheres(et.Json{"key": et.Json{"eq": "CC"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := people.DefineRollup("type_ti", types.Name, fk, []string{"name"}).SetWheres(et.Json{"key": et.Json{"in": []interface{}{"TI", "PA"}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := people.DefineRollup("type_none", types.Name, fk, []string{"name"}).SetWheres(et.Json{"key": et.Json{"eq": "NIT"}}); err != nil {
		t.Fatal(err)
	}

	item, err := people.Select(jdb.KEY, "type_code", "type_cc", "type_ti", "type_none").Where(jdb.KEY).Eq("p1").One()
	if err != nil {
		t.Fatal(err)
	}

	if item.Result.Str("type_cc") != "Cedula" || item.Result.Str("type_ti") != "Tarjeta" {
		t.Errorf("rollup wheres: %v", item.Result)
	}
	// La llave coincide pero el filtro no: sin registro.
	if item.Result["type_none"] != nil {
		t.Errorf("type_none must be nil: %v", item.Result["type_none"])
	}

	if _, err := people.DefineRollup("bad", types.Name, fk, []string{"name"}).SetWheres(et.Json{"key": et.Json{"regex": "C"}}); err == nil {
		t.Error("unknown operator must fail")
	}
	if _, err := people.DefineRollup("bad2", types.Name, fk, []string{"name"}).SetWheres(et.Json{"key": "CC"}); err == nil {
		t.Error("where without operator must fail")
	}

	definition, err := json.Marshal(people.Rollup["type_cc"])
	if err != nil {
		t.Fatal(err)
	}
	loaded := jdb.Rollup{}
	if err := json.Unmarshal(definition, &loaded); err != nil || len(loaded.Wheres) != 1 || loaded.Wheres[0].Json("key").Str("eq") != "CC" {
		t.Errorf("serialized wheres: %s", definition)
	}
}

func TestQueryWhereGroups(t *testing.T) {
	db := connect(t)
	types, _ := newKeyModels(t, db)

	// limit 0: todas las filas (con limit > 1 Query devuelve la fila n, defecto conocido de setLimitTx).
	result, err := types.Query(et.Json{"limit": 0, "where": et.Json{"or": []interface{}{
		et.Json{"key": et.Json{"eq": "CC"}},
		et.Json{"key": et.Json{"eq": "NIT"}},
	}}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Int("count") != 2 {
		t.Errorf("or group must return 2 rows: %v", result)
	}

	result, err = types.Query(et.Json{"limit": 0, "where": et.Json{"and": []interface{}{
		et.Json{"code": et.Json{"eq": "1"}},
		et.Json{"key": et.Json{"eq": "TI"}},
	}}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Int("count") != 1 {
		t.Errorf("and group must return 1 row: %v", result)
	}
}

func TestRunQueryWithRollups(t *testing.T) {
	db := connect(t)
	types, people := newKeyModels(t, db)

	// Rollup suelto: no se agrega al modelo, solo a esta consulta.
	rollup := jdb.NewRollup("type_name", types, map[string]string{"type_code": "code"}, []string{"name", "key"}, jdb.StrToShowRollup("objects"))
	if _, err := rollup.SetWheres(et.Json{"key": et.Json{"eq": "TI"}}); err != nil {
		t.Fatal(err)
	}

	list, err := jdb.RunQuery(people, et.Json{"wheres": []interface{}{et.Json{jdb.KEY: et.Json{"eq": "p1"}}}}, rollup)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Result) != 1 || list.Result[0].Json("type_name").Str("name") != "Tarjeta" {
		t.Fatalf("RunQuery rollup: %v", list.Result)
	}

	if _, ok := people.Rollup["type_name"]; ok {
		t.Error("NewRollup must not change the model")
	}
}

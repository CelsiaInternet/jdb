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

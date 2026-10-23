package postgres

import (
	"strings"
	"testing"

	"github.com/celsiainternet/jdb/jdb"
)

/**
* pathField: campo con ruta JSON dentro de _data, como lo resuelve jdb para "_data.sync_external.<llave>".
* @param key string
* @return *jdb.Field
**/
func pathField(key string) *jdb.Field {
	return &jdb.Field{
		Model:      &jdb.QlFrom{As: "A"},
		Name:       "_data",
		TypeColumn: jdb.TpColumn,
		TypeData:   jdb.TypeDataObject,
		Path:       []string{"sync_external", key},
	}
}

// normalize quita espacios para comparar el SQL sin depender del formato.
var normalize = func(sql string) string { return strings.Join(strings.Fields(sql), "") }

func TestWhereJsonPathSql(t *testing.T) {
	cases := []struct {
		name  string
		where func(where *jdb.QlWhere)
		want  string
	}{
		{"texto", func(where *jdb.QlWhere) { where.Where(pathField("oss_order_id")).Eq("O-2") }, `A._data#>>'{sync_external,oss_order_id}'='O-2'`},
		{"numero", func(where *jdb.QlWhere) { where.Where(pathField("attempts")).More(2) }, `(A._data#>>'{sync_external,attempts}')::numeric>2`},
		{"booleano", func(where *jdb.QlWhere) { where.Where(pathField("done")).Eq(true) }, `(A._data#>>'{sync_external,done}')::boolean=true`},
		{"lista de numeros", func(where *jdb.QlWhere) { where.Where(pathField("attempts")).In(1, 2) }, `(A._data#>>'{sync_external,attempts}')::numeric IN (1, 2)`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			where := &jdb.QlWhere{}
			c.where(where)
			got := whereConditions(where)
			if normalize(got) != normalize(c.want) {
				t.Fatalf("got %s, want %s", got, c.want)
			}
		})
	}
}

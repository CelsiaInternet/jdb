package sqlite

import (
	"database/sql"

	"github.com/celsiainternet/elvis/envar"
	"github.com/celsiainternet/elvis/et"
	jdb "github.com/celsiainternet/jdb/jdb"
	_ "modernc.org/sqlite"
)

type SqlLite struct {
	name      string
	params    et.Json
	connStr   string
	db        *sql.DB
	connected bool
	version   int
}

func newDriver() jdb.Driver {
	return &SqlLite{
		name:      jdb.SqliteDriver,
		params:    et.Json{},
		connected: false,
	}
}

func (s *SqlLite) Name() string {
	return s.name
}

func init() {
	jdb.Register(jdb.SqliteDriver, newDriver, jdb.ConnectParams{
		Id:       envar.GetStr("jdb", "DB_ID"),
		Driver:   jdb.SqliteDriver,
		Name:     envar.GetStr("jdb", "DB_NAME"),
		UserCore: true,
		Debug:    envar.GetBool(true, "DEBUG"),
		Params: et.Json{
			"database": envar.GetStr("jdb", "DB_NAME"),
		},
		Validate: []string{
			"DB_NAME",
		},
	})
}

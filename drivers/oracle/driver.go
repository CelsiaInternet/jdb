package oracle

import (
	"database/sql"

	"github.com/celsiainternet/elvis/envar"
	"github.com/celsiainternet/elvis/et"
	jdb "github.com/celsiainternet/jdb/jdb"
	_ "github.com/lib/pq"
)

type Oracle struct {
	name      string
	params    et.Json
	connStr   string
	db        *sql.DB
	connected bool
	version   int
}

func newDriver() jdb.Driver {
	return &Oracle{
		name:      jdb.OracleDriver,
		params:    et.Json{},
		connected: false,
	}
}

func (s *Oracle) Name() string {
	return s.name
}

func init() {
	jdb.Register(jdb.OracleDriver, newDriver, jdb.ConnectParams{
		Id:     envar.GetStr("jdb", "DB_ID"),
		Driver: jdb.OracleDriver,
		Name:   envar.GetStr("jdb", "DB_NAME"),
		Params: et.Json{
			"database":     envar.GetStr("jdb", "DB_NAME"),
			"host":         envar.GetStr("localhost", "DB_HOST"),
			"port":         envar.GetInt(5432, "DB_PORT"),
			"username":     envar.GetStr("admin", "DB_USER"),
			"password":     envar.GetStr("admin", "DB_PASSWORD"),
			"app":          envar.GetStr("jdb", "APP_NAME"),
			"service_name": envar.GetStr("jdb", "ORA_DB_SERVICE_NAME_ORACLE"),
			"ssl":          envar.GetBool(false, "ORA_DB_SSL_ORACLE"),
			"ssl_verify":   envar.GetBool(false, "ORA_DB_SSL_VERIFY_ORACLE"),
			"version":      envar.GetInt(19, "ORA_DB_VERSION_ORACLE"),
		},
		UserCore: true,
		Validate: []string{
			"DB_NAME",
			"DB_HOST",
			"DB_PORT",
			"DB_USER",
			"DB_PASSWORD",
			"ORA_DB_SERVICE_NAME_ORACLE",
			"ORA_DB_SSL_ORACLE",
			"ORA_DB_SSL_VERIFY_ORACLE",
			"ORA_DB_VERSION_ORACLE",
		},
	})
}

package mysql

import (
	"database/sql"

	"github.com/celsiainternet/elvis/envar"
	"github.com/celsiainternet/elvis/et"
	jdb "github.com/celsiainternet/jdb/jdb"
	_ "github.com/go-sql-driver/mysql"
)

type Mysql struct {
	name      string
	params    et.Json
	connStr   string
	db        *sql.DB
	connected bool
	version   int
}

func newDriver() jdb.Driver {
	return &Mysql{
		name:      jdb.MysqlDriver,
		params:    et.Json{},
		connected: false,
	}
}

func (s *Mysql) Name() string {
	return s.name
}

func init() {
	jdb.Register(jdb.PostgresDriver, newDriver, jdb.ConnectParams{
		Id:       envar.GetStr("jdb", "DB_ID"),
		Driver:   jdb.OracleDriver,
		Name:     envar.GetStr("jdb", "DB_NAME"),
		UserCore: true,
		Debug:    envar.GetBool(true, "DEBUG"),
		Params: et.Json{
			"database": envar.GetStr("jdb", "DB_NAME"),
			"host":     envar.GetStr("localhost", "DB_HOST"),
			"port":     envar.GetInt(3306, "DB_PORT"),
			"username": envar.GetStr("admin", "DB_USER"),
			"password": envar.GetStr("admin", "DB_PASSWORD"),
		},
		Validate: []string{
			"DB_NAME",
			"DB_HOST",
			"DB_PORT",
			"DB_USER",
			"DB_PASSWORD",
		},
	})
}

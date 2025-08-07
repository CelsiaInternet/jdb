package sqlite

import (
	"database/sql"
	"errors"

	"github.com/celsiainternet/elvis/envar"
	"github.com/celsiainternet/elvis/et"
	jdb "github.com/celsiainternet/jdb/jdb"
	_ "modernc.org/sqlite"
)

type Connection struct {
	Database string `json:"database"`
	Version  int    `json:"version"`
}

/**
* Chain
* @return string, error
**/
func (s *Connection) Chain() (string, error) {
	err := s.Validate()
	if err != nil {
		return "", err
	}

	result := s.Database

	return result, nil
}

/**
* ToJson
* @return et.Json
**/
func (s *Connection) ToJson() et.Json {
	return et.Json{
		"database": s.Database,
		"version":  s.Version,
	}
}

/**
* Validate
* @return error
**/
func (s *Connection) Validate() error {
	if s.Database == "" {
		return errors.New("database is required")
	}

	return nil
}

type SqlLite struct {
	db         *sql.DB
	name       string
	version    int
	connected  bool
	connection Connection
}

func newDriver() jdb.Driver {
	return &SqlLite{
		name:      jdb.SqliteDriver,
		connected: false,
		connection: Connection{
			Database: envar.GetStr("jdb", "DB_NAME"),
		},
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
		NodeId:   envar.GetInt(1, "NODE_ID"),
		Debug:    envar.GetBool(false, "DEBUG"),
		Params: &Connection{
			Database: envar.GetStr("jdb", "DB_NAME"),
		},
	})
}

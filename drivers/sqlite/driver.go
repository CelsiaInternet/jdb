package sqlite

import (
	"fmt"

	"github.com/celsiainternet/elvis/envar"
	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/utility"
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
* Load
* @param params et.Json
* @return error
**/
func (s *Connection) Load(params et.Json) error {
	database := params.Str("database")
	if utility.ValidStr(database, 0, []string{}) {
		return fmt.Errorf("database is required")
	}

	version := params.Int("version")
	if version == 0 {
		return fmt.Errorf("version is required")
	}

	s.Database = database
	s.Version = version

	return nil
}

/**
* Validate
* @return error
**/
func (s *Connection) Validate() error {
	if s.Database == "" {
		return fmt.Errorf("database is required")
	}

	return nil
}

type SqlLite struct {
	jdb        *jdb.DB
	name       string
	version    int
	connected  bool
	connection Connection
}

func newDriver(db *jdb.DB) jdb.Driver {
	return &SqlLite{
		jdb:       db,
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

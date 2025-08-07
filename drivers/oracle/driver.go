package oracle

import (
	"database/sql"
	"errors"

	"github.com/celsiainternet/elvis/envar"
	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/strs"
	jdb "github.com/celsiainternet/jdb/jdb"
	_ "github.com/lib/pq"
)

type Connection struct {
	Database string `json:"database"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	App      string `json:"app"`
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

	result := strs.Format(`%s://%s:%s@%s:%d/%s?sslmode=disable&application_name=%s`, jdb.OracleDriver, s.Username, s.Password, s.Host, s.Port, s.Database, s.App)

	return result, nil
}

/**
* ToJson
* @return et.Json
**/
func (s *Connection) ToJson() et.Json {
	return et.Json{
		"database": s.Database,
		"host":     s.Host,
		"port":     s.Port,
		"username": s.Username,
		"password": s.Password,
		"app":      s.App,
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
	if s.Host == "" {
		return errors.New("host is required")
	}
	if s.Port == 0 {
		return errors.New("port is required")
	}
	if s.Username == "" {
		return errors.New("username is required")
	}

	if s.Password == "" {
		return errors.New("password is required")
	}

	if s.App == "" {
		return errors.New("app is required")
	}

	return nil
}

type Oracle struct {
	db         *sql.DB
	name       string
	version    int
	connected  bool
	connection Connection
}

func newDriver() jdb.Driver {
	return &Oracle{
		name:      jdb.OracleDriver,
		connected: false,
		connection: Connection{
			Database: envar.GetStr("jdb", "DB_NAME"),
			Host:     envar.GetStr("localhost", "DB_HOST"),
			Port:     envar.GetInt(5432, "DB_PORT"),
			Username: envar.GetStr("admin", "DB_USER"),
			Password: envar.GetStr("admin", "DB_PASSWORD"),
			Version:  envar.GetInt(19, "DB_VERSION"),
		},
	}
}

func (s *Oracle) Name() string {
	return s.name
}

func init() {
	jdb.Register(jdb.OracleDriver, newDriver, jdb.ConnectParams{
		Id:       envar.GetStr("jdb", "DB_ID"),
		Driver:   jdb.OracleDriver,
		UserCore: true,
		NodeId:   envar.GetInt(1, "NODE_ID"),
		Debug:    envar.GetBool(false, "DEBUG"),
		Params: &Connection{
			Database: envar.GetStr("jdb", "DB_NAME"),
			Host:     envar.GetStr("localhost", "DB_HOST"),
			Port:     envar.GetInt(5432, "DB_PORT"),
			Username: envar.GetStr("admin", "DB_USER"),
			Password: envar.GetStr("admin", "DB_PASSWORD"),
			App:      envar.GetStr("jdb", "APP_NAME"),
			Version:  envar.GetInt(19, "DB_VERSION"),
		},
	})
}

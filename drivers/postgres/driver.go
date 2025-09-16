package postgres

import (
	"fmt"

	"github.com/celsiainternet/elvis/envar"
	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/strs"
	"github.com/celsiainternet/elvis/utility"
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

	result := strs.Format(`%s://%s:%s@%s:%d/%s?sslmode=disable&application_name=%s`, jdb.PostgresDriver, s.Username, s.Password, s.Host, s.Port, s.Database, s.App)

	return result, nil
}

/**
* defaultChain
* @return string, error
**/
func (s *Connection) defaultChain() (string, error) {
	return strs.Format(`%s://%s:%s@%s:%d/%s?sslmode=disable&application_name=%s`, jdb.PostgresDriver, s.Username, s.Password, s.Host, s.Port, "postgres", s.App), nil
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
* Load
* @param params et.Json
* @return error
**/
func (s *Connection) Load(params et.Json) error {
	database := params.Str("database")
	if utility.ValidStr(database, 0, []string{}) {
		return fmt.Errorf("database is required")
	}

	host := params.Str("host")
	if utility.ValidStr(host, 0, []string{}) {
		return fmt.Errorf("host is required")
	}

	port := params.Int("port")
	if port == 0 {
		return fmt.Errorf("port is required")
	}

	username := params.Str("username")
	if utility.ValidStr(username, 0, []string{}) {
		return fmt.Errorf("username is required")
	}

	password := params.Str("password")
	if utility.ValidStr(password, 0, []string{}) {
		return fmt.Errorf("password is required")
	}

	app := params.Str("app")
	if utility.ValidStr(app, 0, []string{}) {
		return fmt.Errorf("app is required")
	}

	version := params.Int("version")
	if version == 0 {
		return fmt.Errorf("version is required")
	}

	s.Database = database
	s.Host = host
	s.Port = port
	s.Username = username
	s.Password = password
	s.App = app
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
	if s.Host == "" {
		return fmt.Errorf("host is required")
	}
	if s.Port == 0 {
		return fmt.Errorf("port is required")
	}
	if s.Username == "" {
		return fmt.Errorf("username is required")
	}

	if s.Password == "" {
		return fmt.Errorf("password is required")
	}

	if s.App == "" {
		return fmt.Errorf("app is required")
	}

	return nil
}

type Postgres struct {
	jdb        *jdb.DB
	name       string
	version    int
	connected  bool
	connection Connection
}

func newDriver(db *jdb.DB) jdb.Driver {
	return &Postgres{
		jdb:  db,
		name: jdb.PostgresDriver,
		connection: Connection{
			Database: envar.GetStr("test", "DB_NAME"),
			Host:     envar.GetStr("localhost", "DB_HOST"),
			Port:     envar.GetInt(5432, "DB_PORT"),
			Username: envar.GetStr("admin", "DB_USER"),
			Password: envar.GetStr("admin", "DB_PASSWORD"),
			App:      envar.GetStr("jdb", "APP_NAME"),
			Version:  envar.GetInt(13, "DB_VERSION"),
		},
	}
}

func (s *Postgres) Name() string {
	return s.name
}

func init() {
	jdb.Register(jdb.PostgresDriver, newDriver, jdb.ConnectParams{
		Id:       envar.GetStr("jdb", "DB_ID"),
		Driver:   jdb.PostgresDriver,
		Name:     envar.GetStr("jdb", "DB_NAME"),
		UserCore: true,
		NodeId:   envar.GetInt(0, "NODE_ID"),
		Debug:    envar.GetBool(false, "DEBUG"),
		Params: &Connection{
			Database: envar.GetStr("jdb", "DB_NAME"),
			Host:     envar.GetStr("localhost", "DB_HOST"),
			Port:     envar.GetInt(5432, "DB_PORT"),
			Username: envar.GetStr("admin", "DB_USER"),
			Password: envar.GetStr("admin", "DB_PASSWORD"),
			App:      envar.GetStr("jdb", "APP_NAME"),
			Version:  envar.GetInt(13, "DB_VERSION"),
		},
	})
}

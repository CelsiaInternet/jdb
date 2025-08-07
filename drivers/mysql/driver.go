package mysql

import (
	"database/sql"
	"errors"

	"github.com/celsiainternet/elvis/envar"
	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/strs"
	jdb "github.com/celsiainternet/jdb/jdb"
	_ "github.com/go-sql-driver/mysql"
)

type Connection struct {
	Database string `json:"database"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
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

	result := strs.Format(`%s:%s@tcp(%s:%d)/%s?parseTime=true`, s.Username, s.Password, s.Host, s.Port, s.Database)

	return result, nil
}

/**
* defaultChain
* @return string, error
**/
func (s *Connection) defaultChain() (string, error) {
	return strs.Format(`%s:%s@tcp(%s:%d)/%s?parseTime=true`, s.Username, s.Password, s.Host, s.Port, "mysql"), nil
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

	return nil
}

type Mysql struct {
	db         *sql.DB
	name       string
	version    int
	connected  bool
	connection Connection
}

func newDriver() jdb.Driver {
	return &Mysql{
		name:      jdb.MysqlDriver,
		connected: false,
		connection: Connection{
			Database: envar.GetStr("jdb", "DB_NAME"),
			Host:     envar.GetStr("localhost", "DB_HOST"),
			Port:     envar.GetInt(3306, "DB_PORT"),
			Username: envar.GetStr("admin", "DB_USER"),
			Password: envar.GetStr("admin", "DB_PASSWORD"),
			Version:  envar.GetInt(8, "DB_VERSION"),
		},
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
		NodeId:   envar.GetInt(1, "NODE_ID"),
		Debug:    envar.GetBool(false, "DEBUG"),
		Params: &Connection{
			Database: envar.GetStr("jdb", "DB_NAME"),
			Host:     envar.GetStr("localhost", "DB_HOST"),
			Port:     envar.GetInt(3306, "DB_PORT"),
			Username: envar.GetStr("admin", "DB_USER"),
			Password: envar.GetStr("admin", "DB_PASSWORD"),
			Version:  envar.GetInt(8, "DB_VERSION"),
		},
	})
}

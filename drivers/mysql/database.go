package mysql

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/celsiainternet/elvis/console"
	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/strs"
	"github.com/celsiainternet/elvis/utility"
	jdb "github.com/celsiainternet/jdb/jdb"
)

/**
* chain
* @param params et.Json
* @return string, error
**/
func (s *Mysql) chain(params et.Json) (string, error) {
	username := params.Str("username")
	password := params.Str("password")
	host := params.Str("host")
	port := params.Int("port")
	database := params.Str("database")

	if !utility.ValidStr(username, 0, []string{""}) {
		return "", fmt.Errorf(jdb.MSS_PARAM_REQUIRED, "username")
	}

	if !utility.ValidStr(password, 0, []string{""}) {
		return "", fmt.Errorf(jdb.MSS_PARAM_REQUIRED, "password")
	}

	if !utility.ValidStr(host, 0, []string{""}) {
		return "", fmt.Errorf(jdb.MSS_PARAM_REQUIRED, "host")
	}

	if port == 0 {
		return "", fmt.Errorf(jdb.MSS_PARAM_REQUIRED, "port")
	}

	if !utility.ValidStr(database, 0, []string{""}) {
		return "", fmt.Errorf(jdb.MSS_PARAM_REQUIRED, "database")
	}

	result := strs.Format(`%s:%s@tcp(%s:%d)/%s?parseTime=true`, username, password, host, port, database)

	return result, nil
}

/**
* connectTo
* @param connStr string
* @return *sql.DB, error
**/
func (s *Mysql) connectTo(connStr string) (*sql.DB, error) {
	db, err := sql.Open(s.name, connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

/**
* connect
* @param params et.Json
* @return error
**/
func (s *Mysql) connect(params et.Json) error {
	connStr, err := s.chain(params)
	if err != nil {
		return err
	}

	db, err := s.connectTo(connStr)
	if err != nil {
		return err
	}

	s.db = db
	s.params = params
	s.connStr = connStr
	s.connected = s.db != nil
	s.getVersion()

	return nil
}

/**
* connectDefault
* @param params et.Json
* @return error
**/
func (s *Mysql) connectDefault(params et.Json) error {
	params["database"] = "Mysql"
	return s.connect(params)
}

/**
* existDatabase
* @param name string
* @return bool, error
**/
func (s *Mysql) existDatabase(name string) (bool, error) {
	sql := jdb.SQLDDL(`
	SELECT SCHEMA_NAME 
	FROM INFORMATION_SCHEMA.SCHEMATA 
	WHERE SCHEMA_NAME = $1`, name)
	items, err := jdb.Query(s.db, sql, name)
	if err != nil {
		return false, err
	}

	if items.Count == 0 {
		return false, nil
	}

	return items.Bool(0, "exists"), nil
}

/**
* createDatabase
* @param name string
* @return error
**/
func (s *Mysql) createDatabase(name string) error {
	if s.db == nil {
		return fmt.Errorf(jdb.MSG_NOT_DRIVER_DB)
	}

	exist, err := s.existDatabase(name)
	if err != nil {
		return err
	}

	if exist {
		return nil
	}

	sql := jdb.SQLDDL(`	
	CREATE DATABASE $1`, name)
	_, err = jdb.Exec(s.db, sql, name)
	if err != nil {
		return err
	}

	console.LogF(s.name, `Database %s created`, name)

	return nil
}

/**
* DropDatabase
* @param name string
* @return error
**/
func (s *Mysql) DropDatabase(name string) error {
	if s.db == nil {
		return fmt.Errorf(jdb.MSG_NOT_DRIVER_DB)
	}

	exist, err := s.existDatabase(name)
	if err != nil {
		return err
	}

	if !exist {
		return nil
	}

	sql := jdb.SQLDDL(`DROP DATABASE $1`, name)
	_, err = jdb.Exec(s.db, sql, name)
	if err != nil {
		return err
	}

	console.LogF(s.name, `Database %s droped`, name)

	return nil
}

/**
* getVersion
* @return int
**/
func (s *Mysql) getVersion() int {
	if !s.connected {
		return 0
	}

	if s.db == nil {
		return 0
	}

	if s.version != 0 {
		return s.version
	}

	var version string
	err := s.db.QueryRow("SELECT version();").Scan(&version)
	if err != nil {
		return 0
	}

	split := strings.Split(version, ".")
	v, err := strconv.Atoi(split[0])
	if err != nil {
		v = 0
	}

	if v < 8 {
		console.Alert(fmt.Sprintf(MSG_VERSION_NOT_SUPPORTED, version))
	}

	s.version = v

	return s.version
}

/**
* Connect
* @param params et.Json
* @return error
**/
func (s *Mysql) Connect(params et.Json) (*sql.DB, error) {
	database := params.Str("database")
	err := s.connectDefault(params)
	if err != nil {
		return nil, err
	}

	err = s.createDatabase(database)
	if err != nil {
		return nil, err
	}

	params["database"] = database
	err = s.connect(params)
	if err != nil {
		return nil, err
	}

	console.LogF(s.name, `Connected to %s:%s`, params.Str("host"), database)

	return s.db, nil
}

/**
* Disconnect
* @return error
**/
func (s *Mysql) Disconnect() error {
	if !s.connected {
		return nil
	}

	if s.db != nil {
		s.db.Close()
	}

	return nil
}

package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/celsiainternet/elvis/console"
	"github.com/celsiainternet/elvis/et"
)

func (s *SqlLite) connectTo(connStr string) (*sql.DB, error) {
	if !strings.HasSuffix(connStr, ".db") {
		connStr = connStr + ".db"
	}

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
* Connect
* @param params et.Json
* @return error
**/
func (s *SqlLite) Connect(params et.Json) (*sql.DB, error) {
	database := params.Str("database")
	if database == "" {
		return nil, errors.New("database is required")
	}

	db, err := s.connectTo(database)
	if err != nil {
		return nil, err
	}

	s.db = db
	s.params = params
	s.connStr = database
	s.connected = s.db != nil
	s.getVersion()

	console.LogF(s.name, `Connected to %s:%s`, params.Str("host"), database)

	return s.db, nil
}

/**
* Disconnect
* @return error
**/
func (s *SqlLite) Disconnect() error {
	if !s.connected {
		return nil
	}

	if s.db != nil {
		s.db.Close()
	}

	return nil
}

/**
* getVersion
* @return int
**/
func (s *SqlLite) getVersion() int {
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
	err := s.db.QueryRow("SELECT sqlite_version();").Scan(&version)
	if err != nil {
		return 0
	}

	split := strings.Split(version, ".")
	v, err := strconv.Atoi(split[0])
	if err != nil {
		v = 0
	}

	if v < 3 {
		console.Alert(fmt.Sprintf(MSG_VERSION_NOT_SUPPORTED, version))
	}

	s.version = v

	return s.version
}

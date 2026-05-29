package sqlite

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/celsiainternet/elvis/console"
	jdb "github.com/celsiainternet/jdb/jdb"
)

func (s *SqlLite) connectTo(database string) (*sql.DB, error) {
	if !strings.HasSuffix(database, ".db") {
		database = database + ".db"
	}

	db, err := sql.Open(s.name, database)
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
* @param connection jdb.ConnectParams
* @return error
**/
func (s *SqlLite) Connect(connection jdb.ConnectParams) (*sql.DB, error) {
	database := connection.Params.(*Connection).Database
	if database == "" {
		return nil, fmt.Errorf("database is required")
	}

	db, err := s.connectTo(database)
	if err != nil {
		return nil, err
	}

	// maxOpenConns := envar.GetInt(3, "DB_MAX_OPEN_CONNS")
	// maxIdleConns := envar.GetInt(2, "DB_MAX_IDLE_CONNS")
	// connMaxLifetime := time.Duration(envar.GetInt(30, "DB_CONN_MAX_LIFETIME")) * time.Minute
	// connMaxIdleTime := time.Duration(envar.GetInt(5, "DB_CONN_MAX_IDLE_TIME")) * time.Minute
	// db.SetMaxOpenConns(maxOpenConns)
	// db.SetMaxIdleConns(maxIdleConns)
	// db.SetConnMaxLifetime(connMaxLifetime)
	// db.SetConnMaxIdleTime(connMaxIdleTime)

	s.connected = db != nil
	console.LogKF(s.name, `Connected to %s`, database)

	return db, nil
}

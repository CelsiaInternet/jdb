package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/celsiainternet/elvis/console"
	"github.com/celsiainternet/elvis/envar"
	jdb "github.com/celsiainternet/jdb/jdb"
)

/**
* Connect
* @param connection jdb.ConnectParams
* @return *sql.DB, error
**/
func (s *Params) Connect(connection jdb.ConnectParams) (*sql.DB, error) {
	if s.jdb == nil {
		return nil, fmt.Errorf(MSG_JDB_NOT_DEFINED)
	}

	params, ok := connection.Params.(*Connection)
	if !ok || params == nil {
		params = &Connection{}
	}
	if params.Database == "" {
		params.Database = connection.Name + ".db"
	}

	chain, err := params.Chain()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", chain)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	if params.Database == ":memory:" {
		// Cada conexión abriría una base en memoria distinta
		db.SetMaxOpenConns(1)
	} else {
		db.SetMaxOpenConns(envar.GetInt(3, "DB_MAX_OPEN_CONNS"))
		db.SetMaxIdleConns(envar.GetInt(1, "DB_MAX_IDLE_CONNS"))
		db.SetConnMaxLifetime(time.Duration(envar.GetInt(30, "DB_CONN_MAX_LIFETIME")) * time.Minute)
		db.SetConnMaxIdleTime(time.Duration(envar.GetInt(5, "DB_CONN_MAX_IDLE_TIME")) * time.Minute)
	}

	s.connection = *params
	s.connected = true
	console.LogKF(s.name, `Connected to %s`, params.Database)

	return db, nil
}

/**
* exec: SQLite no ejecuta varias sentencias en un solo Query, se lanzan una a una.
* @param isDebug bool, statements ...string
* @return error
**/
func (s *Params) exec(isDebug bool, statements ...string) error {
	if s.jdb == nil || s.jdb.Db == nil {
		return fmt.Errorf(MSG_JDB_NOT_DEFINED)
	}

	for _, sql := range statements {
		if sql == "" {
			continue
		}

		if isDebug {
			console.Debug(sql)
		}

		_, err := s.jdb.Db.Exec(sql)
		if err != nil {
			return fmt.Errorf("%w: %s", err, sql)
		}
	}

	return nil
}

/**
* execIsolated: ejecuta las sentencias en una transacción sobre una sola conexión con foreign_keys
* desactivado y legacy_alter_table activo (los PRAGMA no tienen efecto dentro de una transacción).
* @param isDebug bool, statements ...string
* @return error
**/
func (s *Params) execIsolated(isDebug bool, statements ...string) error {
	if s.jdb == nil || s.jdb.Db == nil {
		return fmt.Errorf(MSG_JDB_NOT_DEFINED)
	}

	ctx := context.Background()
	conn, err := s.jdb.Db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.ExecContext(ctx, "PRAGMA foreign_keys = OFF")
	if err != nil {
		return err
	}
	defer conn.ExecContext(ctx, "PRAGMA foreign_keys = ON")

	_, err = conn.ExecContext(ctx, "PRAGMA legacy_alter_table = ON")
	if err != nil {
		return err
	}
	defer conn.ExecContext(ctx, "PRAGMA legacy_alter_table = OFF")

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	for _, sql := range statements {
		if sql == "" {
			continue
		}

		if isDebug {
			console.Debug(sql)
		}

		_, err := tx.ExecContext(ctx, sql)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("%w: %s", err, sql)
		}
	}

	return tx.Commit()
}

/**
* existTable
* @param model *jdb.Model
* @return bool, error
**/
func (s *Params) existTable(model *jdb.Model) (bool, error) {
	sql := `SELECT COUNT(*) AS count FROM sqlite_master WHERE type = 'table' AND name = ?1;`
	items, err := jdb.Query(s.jdb, sql, tableBase(model))
	if err != nil {
		return false, err
	}

	if items.Count == 0 {
		return false, nil
	}

	return items.Int(0, "count") > 0, nil
}

/**
* tableColumns: columnas que hoy tiene la tabla en la base.
* @param model *jdb.Model
* @return map[string]bool, error
**/
func (s *Params) tableColumns(model *jdb.Model) (map[string]bool, error) {
	sql := `SELECT name FROM pragma_table_info(?1);`
	items, err := jdb.Query(s.jdb, sql, tableBase(model))
	if err != nil {
		return nil, err
	}

	result := map[string]bool{}
	for _, item := range items.Result {
		result[item.Str("name")] = true
	}

	return result, nil
}

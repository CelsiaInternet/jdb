package oracle

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/celsiainternet/elvis/console"
	"github.com/celsiainternet/elvis/envar"
	"github.com/celsiainternet/jdb/jdb"
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
		return nil, fmt.Errorf("oracle connection params are required")
	}
	if connection.HostName != "" {
		params.Host = connection.HostName
	}

	chain, err := params.Chain()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("oracle", chain)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(envar.GetInt(3, "DB_MAX_OPEN_CONNS"))
	db.SetMaxIdleConns(envar.GetInt(1, "DB_MAX_IDLE_CONNS"))
	db.SetConnMaxLifetime(time.Duration(envar.GetInt(30, "DB_CONN_MAX_LIFETIME")) * time.Minute)
	db.SetConnMaxIdleTime(time.Duration(envar.GetInt(5, "DB_CONN_MAX_IDLE_TIME")) * time.Minute)

	if params.Version > 0 {
		s.version = params.Version
	}
	s.connection = *params
	s.connected = true
	console.LogKF(s.name, `Connected to %s:%d/%s`, params.Host, params.Port, params.ServiceName)

	return db, nil
}

/**
* exec: Oracle no acepta varias sentencias en una llamada ni el ";" final.
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
* execIndex: como exec, pero ignora ORA-00955 (el nombre ya existe) y ORA-01408 (esas columnas ya
* tienen índice, p. ej. por la llave primaria): Oracle 19c no tiene CREATE INDEX IF NOT EXISTS.
* @param isDebug bool, statements ...string
* @return error
**/
func (s *Params) execIndex(isDebug bool, statements ...string) error {
	for _, sql := range statements {
		err := s.exec(isDebug, sql)
		if err != nil && !isOraError(err, "ORA-00955", "ORA-01408", "ORA-02260", "ORA-02261", "ORA-02275") {
			return err
		}
	}

	return nil
}

/**
* isOraError
* @param err error, codes ...string
* @return bool
**/
func isOraError(err error, codes ...string) bool {
	if err == nil {
		return false
	}

	msg := err.Error()
	for _, code := range codes {
		if strings.Contains(msg, code) {
			return true
		}
	}

	return false
}

/**
* existTable
* @param name string
* @return bool, error
**/
func (s *Params) existTable(name string) (bool, error) {
	items, err := jdb.Query(s.jdb, `SELECT COUNT(*) AS "count" FROM user_tables WHERE table_name = :1`, name)
	if err != nil {
		return false, err
	}

	if items.Count == 0 {
		return false, nil
	}

	return toInt(items.Result[0]["count"]) > 0, nil
}

/**
* dropTable: Oracle 19c no tiene DROP TABLE IF EXISTS.
* @param isDebug bool, name string
* @return error
**/
func (s *Params) dropTable(isDebug bool, name string) error {
	exist, err := s.existTable(name)
	if err != nil {
		return err
	}

	if !exist {
		return nil
	}

	return s.exec(isDebug, s.ddlTableDrop(quoteIdent(name)))
}

/**
* tableColumns: columnas que hoy tiene la tabla en la base.
* @param name string
* @return map[string]bool, error
**/
func (s *Params) tableColumns(name string) (map[string]bool, error) {
	items, err := jdb.Query(s.jdb, `SELECT column_name AS "name" FROM user_tab_columns WHERE table_name = :1`, name)
	if err != nil {
		return nil, err
	}

	result := map[string]bool{}
	for _, item := range items.Result {
		result[item.Str("name")] = true
	}

	return result, nil
}

package postgres

import (
	"fmt"

	"github.com/celsiainternet/elvis/console"
	jdb "github.com/celsiainternet/jdb/jdb"
)

/**
* loadSchema
* @param name string
* @return error
**/
func (s *Postgres) loadSchema(name string) error {
	if s.jdb == nil {
		return fmt.Errorf(MSG_JDB_NOT_DEFINED)
	}

	exist, err := s.existSchema(name)
	if err != nil {
		return err
	}

	if exist {
		return nil
	}

	sql := jdb.SQLDDL(`CREATE SCHEMA IF NOT EXISTS $1`, name)
	err = jdb.Definition(s.jdb, sql)
	if err != nil {
		return err
	}

	console.LogKF(s.name, `Schema %s created`, name)

	return nil
}

/**
* DropSchema
* @param name string
* @return error
**/
func (s *Postgres) DropSchema(name string) error {
	if s.jdb == nil {
		return fmt.Errorf(MSG_JDB_NOT_DEFINED)
	}

	sql := jdb.SQLDDL(`DROP SCHEMA IF EXISTS $1 CASCADE`, name)
	err := jdb.Definition(s.jdb, sql)
	if err != nil {
		return err
	}

	console.LogKF(s.name, `Schema %s droped`, name)

	return nil
}

/**
* existSchema
* @param name string
* @return bool, error
**/
func (s *Postgres) existSchema(name string) (bool, error) {
	if s.jdb == nil {
		return false, fmt.Errorf(MSG_JDB_NOT_DEFINED)
	}

	items, err := jdb.Query(s.jdb, `
	SELECT EXISTS(
		SELECT 1
		FROM information_schema.schemata
		WHERE UPPER(schema_name) = UPPER($1)
	);`, name)
	if err != nil {
		return false, err
	}

	if items.Count == 0 {
		return false, nil
	}

	return items.Bool(0, "exists"), nil
}

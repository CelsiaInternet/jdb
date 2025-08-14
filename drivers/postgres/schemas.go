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
	if s.db == nil {
		return fmt.Errorf(jdb.MSG_NOT_DRIVER_DB)
	}

	exist, err := s.existSchema(name)
	if err != nil {
		return err
	}

	if exist {
		return nil
	}

	sql := jdb.SQLDDL(`CREATE SCHEMA IF NOT EXISTS $1`, name)
	_, err = jdb.Exec(s.db, sql)
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
	if s.db == nil {
		return fmt.Errorf(jdb.MSG_NOT_DRIVER_DB)
	}

	sql := jdb.SQLDDL(`DROP SCHEMA IF EXISTS $1 CASCADE`, name)
	_, err := jdb.Query(s.db, sql)
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
	if s.db == nil {
		return false, fmt.Errorf(jdb.MSG_NOT_DRIVER_DB)
	}

	sql := jdb.SQLDDL(`SELECT 1 FROM pg_namespace WHERE nspname = '$1';`, name)
	items, err := jdb.Query(s.db, sql)
	if err != nil {
		return false, err
	}

	return items.Ok, nil
}

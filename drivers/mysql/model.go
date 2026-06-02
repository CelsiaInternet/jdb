package mysql

import (
	"fmt"

	"github.com/celsiainternet/elvis/console"
	"github.com/celsiainternet/elvis/strs"
	jdb "github.com/celsiainternet/jdb/jdb"
)

func tableName(model *jdb.Model) string {
	return fmt.Sprintf(`%s_%s`, model.Schema, model.Table)
}

/**
* existTable
* @param db, name string
* @return bool, error
**/
func (s *Mysql) existTable(db, name string) (bool, error) {
	sql := `
	SELECT EXISTS(
		SELECT 1
		FROM information_schema.tables
		WHERE UPPER(table_schema) = UPPER($1)
		AND UPPER(table_name) = UPPER($2));`

	items, err := jdb.Query(s.jdb, sql, db, name)
	if err != nil {
		return false, err
	}

	if items.Count == 0 {
		return false, nil
	}

	return items.Bool(0, "exists"), nil
}

/**
* LoadModel
* @param model *jdb.Model
* @return (bool, error)
**/
func (s *Mysql) LoadModel(model *jdb.Model) (bool, error) {
	table := tableName(model)
	exist, err := s.existTable(model.Db.Name, table)
	if err != nil {
		return false, err
	}

	if exist {
		return false, nil
	}

	sql := s.ddlTable(model)
	sqlIndex := s.ddlTableIndex(model)
	sql = strs.Append(sql, sqlIndex, "\n")
	if model.IsDebug {
		console.Debug(sql)
	}

	_, err = jdb.Query(s.jdb, sql)
	if err != nil {
		return false, err
	}

	return true, nil
}

/**
* DropModel
* @param model *jdb.Model
* @return error
**/
func (s *Mysql) DropModel(model *jdb.Model) error {
	sql := s.ddlTableDrop(tableName(model))
	if model.IsDebug {
		console.Debug(sql)
	}

	_, err := jdb.Query(s.jdb, sql)
	if err != nil {
		return err
	}

	return nil
}

/**
* EmptyModel
* @param model *jdb.Model
* @return error
**/
func (s *Mysql) EmptyModel(model *jdb.Model) error {
	sql := s.ddlTableEmpty(tableName(model))
	if model.IsDebug {
		console.Debug(sql)
	}

	_, err := jdb.Query(s.jdb, sql)
	if err != nil {
		return err
	}

	return nil
}

/**
* MutateModel
* @param model *jdb.Model
* @return error
**/
func (s *Mysql) MutateModel(model *jdb.Model) error {
	backupTable := strs.Format(`%s_backup`, tableName(model))
	sql := "\n"
	sql = strs.Append(sql, s.ddlTableRename(tableName(model), backupTable), "\n")
	sql = strs.Append(sql, s.ddlTable(model), "\n")
	sql = strs.Append(sql, s.ddlTableInsertTo(model, backupTable), "\n\n")
	sql = strs.Append(sql, s.ddlTableIndex(model), "\n\n")
	sql = strs.Append(sql, s.ddlTableDrop(backupTable), "\n\n")
	if model.IsDebug {
		console.Debug(sql)
	}

	_, err := jdb.Query(s.jdb, sql)
	if err != nil {
		return err
	}

	return nil
}

package postgres

import (
	"fmt"

	"github.com/celsiainternet/elvis/console"
	"github.com/celsiainternet/elvis/strs"
	jdb "github.com/celsiainternet/jdb/jdb"
)

func tableName(model *jdb.Model) string {
	return fmt.Sprintf(`%s.%s`, model.Schema, model.Table)
}

/**
* existTable
* @param schema, name string
* @return bool, error
**/
func (s *Postgres) existTable(schema, name string) (bool, error) {
	sql := `
	SELECT EXISTS(
		SELECT 1
		FROM information_schema.tables
		WHERE UPPER(table_schema) = UPPER($1)
		AND UPPER(table_name) = UPPER($2));`
	items, err := jdb.Query(s.jdb, sql, schema, name)
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
func (s *Postgres) LoadModel(model *jdb.Model) (bool, error) {
	err := s.loadSchema(model.Schema)
	if err != nil {
		return false, err
	}

	exist, err := s.existTable(model.Schema, model.Table)
	if err != nil {
		return false, err
	}

	if exist {
		return true, nil
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

	console.LogKF("Model", "Create %s", tableName(model))

	return false, nil
}

/**
* DropModel
* @param model *jdb.Model
* @return error
**/
func (s *Postgres) DropModel(model *jdb.Model) error {
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
func (s *Postgres) EmptyModel(model *jdb.Model) error {
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
func (s *Postgres) MutateModel(model *jdb.Model) error {
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

package jdb

import (
	"database/sql"
	"fmt"

	"github.com/celsiainternet/elvis/console"
	"github.com/celsiainternet/elvis/et"
)

/**
* QueryTx
* @param tx *Tx, db *sql.DB, sql string, arg ...any
* @return *sql.Rows, error
**/
func QueryTx(tx *Tx, db *sql.DB, sql string, arg ...any) (et.Items, error) {
	if tx != nil {
		err := tx.Begin(db)
		if err != nil {
			return et.Items{}, err
		}

		rows, err := tx.Tx.Query(sql, arg...)
		if err != nil {
			sql = SQLParse(sql, arg...)
			errRollback := tx.Rollback()
			if errRollback != nil {
				err = fmt.Errorf("error on rollback: %w: %s", errRollback, err)
			}

			return et.Items{}, console.Alert(fmt.Sprintf("QueryTx error: %s", err.Error()))
		}
		defer rows.Close()

		result := RowsToItems(rows)

		return result, nil
	}

	rows, err := db.Query(sql, arg...)
	if err != nil {
		sql = SQLParse(sql, arg...)
		return et.Items{}, console.Alert(fmt.Sprintf("Query error: %s", err.Error()))
	}
	defer rows.Close()

	result := RowsToItems(rows)

	return result, nil
}

/**
* Query
* @param db *sql.DB, sql string, arg ...any
* @return et.Items, error
**/
func Query(db *sql.DB, sql string, arg ...any) (et.Items, error) {
	return QueryTx(nil, db, sql, arg...)
}

/**
* Exec
* @param db *sql.DB, sql string, arg ...any
* @return et.Items, error
**/
func Exec(db *sql.DB, sql string, arg ...any) (et.Items, error) {
	result, err := Query(db, sql, arg...)
	if err != nil {
		return et.Items{}, err
	}

	sql = SQLParse(sql, arg...)
	audit("exec", sql)

	return result, nil
}

/**
* DataTx
* @param tx *Tx, db *sql.DB, sourceFiled, sql string, arg ...any
* @return et.Items, error
**/
func DataTx(tx *Tx, db *sql.DB, sourceFiled, sql string, arg ...any) (et.Items, error) {
	if tx != nil {
		err := tx.Begin(db)
		if err != nil {
			return et.Items{}, err
		}

		rows, err := tx.Tx.Query(sql, arg...)
		if err != nil {
			errRollback := tx.Rollback()
			if errRollback != nil {
				err = fmt.Errorf("error on rollback: %w: %s", errRollback, err)
			}

			return et.Items{}, console.Alert(fmt.Sprintf("DataTx error: %s", err.Error()))
		}
		defer rows.Close()

		result := RowsToSource(sourceFiled, rows)

		return result, nil
	}

	rows, err := db.Query(sql, arg...)
	if err != nil {
		sql = SQLParse(sql, arg...)
		return et.Items{}, console.Alert(fmt.Sprintf("Data error: %s", err.Error()))
	}
	defer rows.Close()

	result := RowsToSource(sourceFiled, rows)

	return result, nil
}

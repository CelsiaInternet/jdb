package jdb

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/event"
)

func tipoSQL(query string) string {
	q := strings.TrimSpace(strings.ToUpper(query))

	parts := strings.Fields(q)
	if len(parts) == 0 {
		return "DESCONOCIDO"
	}

	cmd := parts[0]

	switch cmd {
	case "SELECT":
		return "query"
	case "INSERT", "UPDATE", "DELETE", "MERGE":
		return "command"
	case "CREATE", "ALTER", "DROP", "TRUNCATE":
		return "definition"
	case "GRANT", "REVOKE":
		return "control"
	case "COMMIT", "ROLLBACK", "SAVEPOINT", "SET":
		return "transaction"
	default:
		return "desconocido"
	}
}

/**
* RowsToItems
* @param rows *sql.Rows
* @return et.Items
**/
func RowsToItems(rows *sql.Rows) et.Items {
	var result = et.Items{Result: []et.Json{}}

	append := func(item et.Json) {
		result.Add(item)
	}

	for rows.Next() {
		var item et.Json
		item.ScanRows(rows)

		if len(item) == 1 {
			for _, v := range item {
				switch val := v.(type) {
				case et.Json:
					append(val)
				case map[string]interface{}:
					append(et.Json(val))
				default:
					append(item)
				}
			}
		} else {
			append(item)
		}
	}

	return result
}

/**
* queryTx
* @param db *DB, tx *Tx, sql string, arg ...any
* @return *sql.Rows, error
**/
func queryTx(db *DB, tx *Tx, query string, arg ...any) (et.Items, error) {
	data := et.Json{
		"db_name": db.Name,
		"query":   query,
		"args":    arg,
	}

	var err error
	var rows *sql.Rows
	if tx != nil {
		err = tx.Begin(db.Db)
		if err != nil {
			return et.Items{}, err
		}

		rows, err = tx.Tx.Query(query, arg...)
		if err != nil {
			errRollback := tx.Rollback()
			if errRollback != nil {
				data["error"] = err.Error()
				event.Publish(EVENT_SQL_ERROR, data)
				err = fmt.Errorf("error on rollback: %w: %s", errRollback, err)
			}

			return et.Items{}, err
		}
	} else {
		rows, err = db.Db.Query(query, arg...)
		if err != nil {
			return et.Items{}, err
		}
	}

	tp := tipoSQL(query)
	event.Publish(fmt.Sprintf("sql:%s", tp), data)
	defer rows.Close()
	result := RowsToItems(rows)
	return result, nil
}

/**
* QueryTx
* @param db *DB, tx *Tx, sql string, arg ...any
* @return et.Items, error
**/
func QueryTx(db *DB, tx *Tx, sql string, arg ...any) (et.Items, error) {
	return queryTx(db, tx, sql, arg...)
}

/**
* Query
* @param db *DB, sql string, arg ...any
* @return et.Items, error
**/
func Query(db *DB, sql string, arg ...any) (et.Items, error) {
	return QueryTx(db, nil, sql, arg...)
}

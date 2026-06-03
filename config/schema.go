package config

import (
	"github.com/celsiainternet/jdb/jdb"
)

/**
* defineSchema
* @param db *jdb.DB, name string
* @return (*jdb.Schema, error)
**/
func defineSchema(db *jdb.DB, name string) (*jdb.Schema, error) {
	schema := jdb.NewSchema(db, name)
	return schema, nil
}

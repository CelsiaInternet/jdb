package instances

import (
	"github.com/celsiainternet/jdb/jdb"
)

func defineSchema(db *jdb.DB, name string) (*jdb.Schema, error) {
	schema := jdb.NewSchema(db, name)
	return schema, nil
}

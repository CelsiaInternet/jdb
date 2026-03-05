package instances

import (
	"github.com/celsiainternet/jdb/jdb"
)

func (i *Instance) defineSchema(db *jdb.DB, name string) error {
	if i.schema == nil {
		i.schema = jdb.NewSchema(db, name)
	}

	return nil
}

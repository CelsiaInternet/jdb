package postgres

import (
	"github.com/celsiainternet/elvis/console"
	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/strs"
	jdb "github.com/celsiainternet/jdb/jdb"
)

/**
* Command
* @param command *jdb.Command
* @return et.Items, error
**/
func (s *Postgres) Command(command *jdb.Command) (et.Items, error) {
	command.Sql = ""
	switch command.Command {
	case jdb.Insert:
		command.Sql = strs.Append(command.Sql, s.sqlInsert(command), "\n")
	case jdb.Update:
		command.Sql = strs.Append(command.Sql, s.sqlUpdate(command), "\n")
	case jdb.Delete:
		command.Sql = strs.Append(command.Sql, s.sqlDelete(command), "\n")
	}

	if command.IsDebug {
		console.Debug(command.Sql)
	}

	result, err := jdb.QueryTx(s.jdb, command.Tx(), command.Sql)
	if err != nil {
		return et.Items{}, err
	}

	return result, nil
}

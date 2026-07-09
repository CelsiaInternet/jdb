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
	command.Args = []any{}
	switch command.Command {
	case jdb.Insert:
		sql, args := s.sqlInsert(command)
		command.Sql = strs.Append(command.Sql, sql, "\n")
		command.Args = append(command.Args, args...)
	case jdb.Update:
		sql, args := s.sqlUpdate(command)
		command.Sql = strs.Append(command.Sql, sql, "\n")
		command.Args = append(command.Args, args...)
	case jdb.Delete:
		sql, args := s.sqlDelete(command)
		command.Sql = strs.Append(command.Sql, sql, "\n")
		command.Args = append(command.Args, args...)
	}

	if command.IsDebug {
		console.Debug(et.Json{"sql": command.Sql, "args": command.Args}.ToString())
	}

	result, err := jdb.QueryTx(s.jdb, command.Tx(), command.Sql, command.Args...)
	if err != nil {
		console.Error(err)
		console.Debug(command.Sql)
		console.Debug(command.Args)
		return et.Items{}, err
	}

	return result, nil
}

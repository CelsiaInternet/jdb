package jdb

import (
	"github.com/celsiainternet/elvis/console"
	"github.com/celsiainternet/elvis/envar"
	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/utility"
)

var coreAudit *Model

func (s *DB) defineAudit() error {
	if s.driver.Name() == SqliteDriver {
		return nil
	}

	if err := s.defineSchema(); err != nil {
		return err
	}

	if coreAudit != nil {
		return nil
	}

	coreAudit = NewModel(coreSchema, "audit", 1)
	coreAudit.DefineColumn(cf.CreatedAt, TypeDataDateTime)
	coreAudit.DefineColumn("command", TypeDataText)
	coreAudit.DefineColumn("query", TypeDataMemo)
	coreAudit.definePrimaryKeyField()
	coreAudit.DefineIndexField()
	coreAudit.DefineIndex(true,
		cf.CreatedAt,
		"command",
	)
	coreAudit.isAudit = true
	if err := coreAudit.Init(); err != nil {
		return console.Panic(err)
	}

	return nil
}

/**
* audit
* @param command, query string
**/
func audit(command string, query string) {
	if coreAudit == nil || !coreAudit.isInit {
		return
	}

	result := utility.ToBase64(query)
	_, err := coreAudit.Insert(et.Json{
		cf.CreatedAt: utility.Now(),
		cf.Key:       coreAudit.GenId(),
		"command":    command,
		"query":      result,
	}).
		AfterInsert(func(tx *Tx, data et.Json) error {
			count, err := coreAudit.
				Counted()
			if err != nil {
				return err
			}

			limit := envar.GetInt(10000, "AUDIT_LIMIT")
			if count > limit {
				item, err := coreAudit.
					Where("command").Neg("exec").
					OrderBy(cf.Index).
					First(1)
				if err != nil {
					return err
				}

				id := item.Str(0, cf.Key)
				_, err = coreAudit.
					Delete(cf.Key).Eq(id).
					ExecTx(tx)
				if err != nil {
					return err
				}
			}

			return nil
		}).
		Exec()
	if err != nil {
		console.Alert(err.Error())
	}

	debug := envar.GetBool(true, "DEBUG")

	if debug {
		console.Debug("Audit:", query)
	}
}

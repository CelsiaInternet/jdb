package sqlite

import (
	"github.com/celsiainternet/elvis/console"
	"github.com/celsiainternet/elvis/envar"
	"github.com/celsiainternet/elvis/et"
	jdb "github.com/celsiainternet/jdb/jdb"
	_ "modernc.org/sqlite"
)

/**
* Connection: Database es la ruta del archivo (o ":memory:"); si está vacía se usa "<ConnectParams.Name>.db".
**/
type Connection struct {
	Database string `json:"database"`
	UseCore  bool   `json:"use_core"`
	IsDebug  bool   `json:"is_debug"`
}

func init() {
	jdb.Register(jdb.SqliteDriver, newDriver, jdb.ConnectParams{
		Id:      envar.GetStr("jdb", "DB_ID"),
		Driver:  jdb.SqliteDriver,
		Name:    envar.GetStr("jdb", "DB_NAME"),
		UseCore: envar.GetBool(true, "USE_CORE"),
		IsDebug: envar.GetBool(false, "DEBUG"),
		Params: &Connection{
			Database: envar.GetStr("", "SQLITE_PATH"),
			UseCore:  envar.GetBool(true, "USE_CORE"),
			IsDebug:  envar.GetBool(false, "DEBUG"),
		},
	})
}

/**
* Chain: DSN de modernc.org/sqlite con foreign keys, WAL y busy timeout.
* @return string, error
**/
func (s *Connection) Chain() (string, error) {
	err := s.Validate()
	if err != nil {
		return "", err
	}

	if s.Database == ":memory:" {
		return "file::memory:?_pragma=foreign_keys(1)", nil
	}

	return "file:" + s.Database + "?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)", nil
}

/**
* ToJson
* @return et.Json
**/
func (s *Connection) ToJson() et.Json {
	return et.Json{
		"database": s.Database,
		"use_core": s.UseCore,
		"is_debug": s.IsDebug,
	}
}

/**
* Load
* @param params et.Json
* @return error
**/
func (s *Connection) Load(params et.Json) error {
	s.Database = params.Str("database")
	s.UseCore = params.Bool("use_core")
	s.IsDebug = params.Bool("is_debug")

	return nil
}

/**
* Validate: la ruta puede venir vacía, Connect la completa con el nombre de la base.
* @return error
**/
func (s *Connection) Validate() error {
	return nil
}

type Params struct {
	jdb        *jdb.DB
	name       string
	connected  bool
	connection Connection
}

func newDriver(db *jdb.DB) jdb.Driver {
	return &Params{
		jdb:  db,
		name: jdb.SqliteDriver,
		connection: Connection{
			Database: envar.GetStr("", "SQLITE_PATH"),
			IsDebug:  envar.GetBool(false, "DEBUG"),
		},
	}
}

func (s *Params) Name() string {
	return s.name
}

/**
* LoadModel
* @param model *jdb.Model
* @return (bool, error)
**/
func (s *Params) LoadModel(model *jdb.Model) (bool, error) {
	exist, err := s.existTable(model)
	if err != nil {
		return false, err
	}

	if exist {
		return true, nil
	}

	statements := []string{s.ddlTable(model, tableName(model))}
	statements = append(statements, s.ddlTableIndex(model)...)
	err = s.exec(model.IsDebug, statements...)
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
func (s *Params) DropModel(model *jdb.Model) error {
	return s.exec(model.IsDebug, s.ddlTableDrop(tableName(model)))
}

/**
* EmptyModel
* @param model *jdb.Model
* @return error
**/
func (s *Params) EmptyModel(model *jdb.Model) error {
	return s.exec(model.IsDebug, s.ddlTableEmpty(tableName(model)))
}

/**
* MutateModel: recrea la tabla con la definición actual y copia las columnas que ya existían.
* SQLite no permite ALTER de constraints, así que se crea una tabla nueva, se copia, se borra la vieja
* y se renombra; con foreign_keys desactivado para no disparar cascadas ni reescribir referencias.
* @param model *jdb.Model
* @return error
**/
func (s *Params) MutateModel(model *jdb.Model) error {
	table := tableName(model)
	newTable := quoteIdent(tableBase(model) + "_new")
	current, err := s.tableColumns(model)
	if err != nil {
		return err
	}

	statements := []string{
		s.ddlTableDrop(newTable),
		s.ddlTable(model, newTable),
		s.ddlTableInsertTo(model, newTable, table, current),
		s.ddlTableDrop(table),
		s.ddlTableRename(newTable, tableBase(model)),
	}
	statements = append(statements, s.ddlTableIndex(model)...)

	return s.execIsolated(model.IsDebug, statements...)
}

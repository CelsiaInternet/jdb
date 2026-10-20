package oracle

import (
	"errors"

	"github.com/celsiainternet/elvis/console"
	"github.com/celsiainternet/elvis/envar"
	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/jdb/jdb"
	goOra "github.com/sijms/go-ora/v2"
)

type Connection struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	ServiceName string `json:"service_name"`
	SSL         bool   `json:"ssl"`
	SSLVerify   bool   `json:"ssl_verify"`
	Version     int    `json:"version"`
}

func (s *Connection) Chain() (string, error) {
	if err := s.Validate(); err != nil {
		return "", err
	}

	ssl := "false"
	if s.SSL {
		ssl = "true"
	}
	sslVerify := "false"
	if s.SSLVerify {
		sslVerify = "true"
	}
	urlOptions := map[string]string{
		"ssl":        ssl,
		"ssl verify": sslVerify,
	}

	result := goOra.BuildUrl(s.Host, s.Port, s.ServiceName, s.Username, s.Password, urlOptions)
	return result, nil
}

func (s *Connection) ToJson() et.Json {
	return et.Json{
		"host":         s.Host,
		"port":         s.Port,
		"username":     s.Username,
		"password":     s.Password,
		"service_name": s.ServiceName,
		"ssl":          s.SSL,
		"ssl_verify":   s.SSLVerify,
		"version":      s.Version,
	}
}

func (s *Connection) Load(params et.Json) error {
	s.Host = params.Str("host")
	s.Port = params.Int("port")
	s.Username = params.Str("username")
	s.Password = params.Str("password")
	s.ServiceName = params.Str("service_name")
	s.SSL = params.Bool("ssl")
	s.SSLVerify = params.Bool("ssl_verify")
	s.Version = params.Int("version")

	return s.Validate()
}

func (s *Connection) Validate() error {
	if s.Host == "" {
		return errors.New("host is required")
	}
	if s.Port == 0 {
		return errors.New("port is required")
	}
	if s.Username == "" {
		return errors.New("username is required")
	}
	if s.Password == "" {
		return errors.New("password is required")
	}
	if s.ServiceName == "" {
		return errors.New("service_name is required")
	}
	return nil
}

type Params struct {
	jdb        *jdb.DB
	name       string
	version    int
	connected  bool
	connection Connection
}

func init() {
	jdb.Register(jdb.OracleDriver, newDriver, jdb.ConnectParams{
		Id:      envar.GetStr("jdb", "DB_ID"),
		Driver:  jdb.OracleDriver,
		Name:    envar.GetStr("jdb", "DB_NAME"),
		UseCore: envar.GetBool(true, "USE_CORE"),
		IsDebug: envar.GetBool(false, "DEBUG"),
		Params: &Connection{
			Host:        envar.GetStr("", "ORA_DB_HOST"),
			Port:        envar.GetInt(1521, "ORA_DB_PORT"),
			Username:    envar.GetStr("", "ORA_DB_USER"),
			Password:    envar.GetStr("", "ORA_DB_PASSWORD"),
			ServiceName: envar.GetStr("", "ORA_DB_SERVICE_NAME_ORACLE"),
			SSL:         envar.GetBool(false, "ORA_DB_SSL_ORACLE"),
			SSLVerify:   envar.GetBool(false, "ORA_DB_SSL_VERIFY_ORACLE"),
			Version:     envar.GetInt(19, "ORA_VERSION"),
		},
	})
}

/**
* New Driver
* @param db *jdb.DB
* @return jdb.Driver
**/
func newDriver(db *jdb.DB) jdb.Driver {
	return &Params{
		jdb:     db,
		name:    jdb.OracleDriver,
		version: envar.GetInt(19, "ORA_VERSION"),
		connection: Connection{
			Host:        envar.GetStr("", "ORA_DB_HOST"),
			Port:        envar.GetInt(1521, "ORA_DB_PORT"),
			Username:    envar.GetStr("", "ORA_DB_USER"),
			Password:    envar.GetStr("", "ORA_DB_PASSWORD"),
			ServiceName: envar.GetStr("", "ORA_DB_SERVICE_NAME_ORACLE"),
			SSL:         envar.GetBool(false, "ORA_DB_SSL_ORACLE"),
			SSLVerify:   envar.GetBool(false, "ORA_DB_SSL_VERIFY_ORACLE"),
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
	exist, err := s.existTable(tableBase(model))
	if err != nil {
		return false, err
	}

	if exist {
		return true, nil
	}

	err = s.exec(model.IsDebug, s.ddlTable(model, tableName(model)))
	if err != nil {
		return false, err
	}

	err = s.execIndex(model.IsDebug, s.ddlTableIndex(model)...)
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
	return s.dropTable(model.IsDebug, tableBase(model))
}

/**
* EmptyModel: DELETE en vez de TRUNCATE, que Oracle rechaza si la tabla tiene foráneas apuntándole.
* @param model *jdb.Model
* @return error
**/
func (s *Params) EmptyModel(model *jdb.Model) error {
	return s.exec(model.IsDebug, s.ddlTableEmpty(tableName(model)))
}

/**
* MutateModel: crea la tabla nueva, copia las columnas que ya existían, borra la vieja, renombra y
* vuelve a crear llaves e índices. El DDL de Oracle hace commit implícito: no es atómico.
* @param model *jdb.Model
* @return error
**/
func (s *Params) MutateModel(model *jdb.Model) error {
	base := tableBase(model)
	newBase := base + "_new"
	current, err := s.tableColumns(base)
	if err != nil {
		return err
	}

	err = s.dropTable(model.IsDebug, newBase)
	if err != nil {
		return err
	}

	err = s.exec(model.IsDebug,
		s.ddlTable(model, quoteIdent(newBase)),
		s.ddlTableInsertTo(model, quoteIdent(newBase), tableName(model), current),
	)
	if err != nil {
		return err
	}

	err = s.dropTable(model.IsDebug, base)
	if err != nil {
		return err
	}

	err = s.exec(model.IsDebug, s.ddlTableRename(quoteIdent(newBase), base))
	if err != nil {
		return err
	}

	return s.execIndex(model.IsDebug, s.ddlTableIndex(model)...)
}

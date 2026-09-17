package oracle

import (
	"database/sql"
	"errors"
	"fmt"

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
}

func (s *Connection) Chain() (string, error) {
	if err := s.Validate(); err != nil {
		return "", err
	}

	host := s.Host
	port := s.Port
	service_name := s.ServiceName
	user := s.Username
	password := s.Password
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

	result := goOra.BuildUrl(host, port, service_name, user, password, urlOptions)
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
		IsDebug: envar.GetBool(false, "DEBUG"),
		Params: &Connection{
			Host:        envar.GetStr("", "ORA_DB_HOST"),
			Port:        envar.GetInt(1521, "ORA_DB_PORT"),
			Username:    envar.GetStr("", "ORA_DB_USER"),
			Password:    envar.GetStr("", "ORA_DB_PASSWORD"),
			ServiceName: envar.GetStr("", "ORA_DB_SERVICE_NAME_ORACLE"),
			SSL:         envar.GetBool(false, "ORA_DB_SSL_ORACLE"),
			SSLVerify:   envar.GetBool(false, "ORA_DB_SSL_VERIFY_ORACLE"),
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
		version: envar.GetInt(13, "ORA_VERSION"),
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

func (s *Params) Connect(params jdb.ConnectParams) (*sql.DB, error) {
	if s.jdb == nil {
		return nil, fmt.Errorf(MSG_JDB_NOT_DEFINED)
	}

	connStr, err := params.Params.Chain()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("oracle", connStr)
	if err != nil {
		return nil, err
	}

	s.connected = db != nil
	console.LogKF(s.name, `Connected to %s:%s`, params.HostName, params.Name)

	return db, nil
}

func (s *Params) LoadModel(model *jdb.Model) (bool, error) {
	return false, errors.New("not implemented")
}

func (s *Params) DropModel(model *jdb.Model) error {
	return errors.New("not implemented")
}

func (s *Params) EmptyModel(model *jdb.Model) error {
	return errors.New("not implemented")
}

func (s *Params) MutateModel(model *jdb.Model) error {
	return errors.New("not implemented")
}

func (s *Params) Select(ql *jdb.Ql) (et.Items, error) {
	return et.Items{}, errors.New("not implemented")
}

func (s *Params) Count(ql *jdb.Ql) (int, error) {
	return 0, errors.New("not implemented")
}

func (s *Params) Exists(ql *jdb.Ql) (bool, error) {
	return false, errors.New("not implemented")
}

func (s *Params) Command(command *jdb.Command) (et.Items, error) {
	return et.Items{}, errors.New("not implemented")
}

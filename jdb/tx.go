package jdb

import (
	"database/sql"
	"fmt"

	"github.com/celsiainternet/elvis/utility"
)

type Tx struct {
	Id string
	Tx *sql.Tx
}

/**
* newTx: Abre una transaccion sobre la conexion.
* @param db *sql.DB
* @return *Tx, error
**/
func newTx(db *sql.DB) (*Tx, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	return &Tx{
		Id: utility.UUID(),
		Tx: tx,
	}, nil
}

/**
* Commit
* @return error
**/
func (s *Tx) commit() error {
	if s.Tx == nil {
		return nil
	}

	return s.Tx.Commit()
}

/**
* Rollback
* @return error
**/
func (s *Tx) rollback() error {
	if s.Tx == nil {
		return nil
	}

	return s.Tx.Rollback()
}

func (s *Tx) query(query string, args ...interface{}) (*sql.Rows, error) {
	if s.Tx == nil {
		return nil, fmt.Errorf("tx is not open")
	}

	return s.Tx.Query(query, args...)
}

/**
* Begin: Abre una transaccion en la base para agrupar varios comandos: se pasa a los metodos
* ...Tx (ExecTx, OneTx, AllTx, ItExistsTx...) y se cierra con End (o Commit / Rollback). Los
* comandos que reciben una transaccion externa no la confirman ni la revierten.
* @return *Tx, error
**/
func (s *DB) Begin() (*Tx, error) {
	if s == nil || s.Db == nil {
		return nil, fmt.Errorf(MSG_DATABASE_NOT_CONNECTED)
	}

	return newTx(s.Db)
}

/**
* Commit: Confirma la transaccion.
* @return error
**/
func (s *Tx) Commit() error {
	return s.commit()
}

/**
* Rollback: Revierte la transaccion.
* @return error
**/
func (s *Tx) Rollback() error {
	return s.rollback()
}

/**
* End: Cierra la transaccion segun el resultado: con err la revierte y devuelve err; sin err la
* confirma y devuelve el error del commit. Uso: defer o al final, return tx.End(err).
* @param err error
* @return error
**/
func (s *Tx) End(err error) error {
	if err != nil {
		s.rollback()
		return err
	}

	return s.commit()
}

/**
* finish: Para un defer con retorno con nombre: revierte si *err no es nil; si no, confirma y deja
* en *err el error del commit. Asi una transaccion propia nunca queda abierta.
* @param err *error
**/
func (s *Tx) finish(err *error) {
	if *err != nil {
		s.rollback()
		return
	}

	if commitErr := s.commit(); commitErr != nil {
		*err = commitErr
	}
}

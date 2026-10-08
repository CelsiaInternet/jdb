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
* NewTx
* @return *Tx
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

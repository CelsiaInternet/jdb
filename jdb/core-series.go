package jdb

import (
	"fmt"

	"github.com/celsiainternet/elvis/et"
)

var coreSeries *Model

func (s *DB) defineSeries() error {
	if err := s.defineSchema(); err != nil {
		return err
	}

	if coreSeries != nil {
		return nil
	}

	coreSeries = NewModel(coreSchema, "series", 1)
	coreSeries.DefineColumn(cf.CreatedAt, TypeDataDateTime)
	coreSeries.DefineColumn(cf.UpdatedAt, TypeDataDateTime)
	coreSeries.DefineColumn("kind", TypeDataText)
	coreSeries.DefineColumn("tag", TypeDataText)
	coreSeries.DefineColumn("value", TypeDataInt)
	coreSeries.DefineColumn("format", TypeDataText)
	coreSeries.DefineSystemKeyField()
	coreSeries.DefinePrimaryKey("kind", "tag")
	coreSeries.DefineIndex(true,
		"format",
		cf.SystemId,
	)
	if err := coreSeries.Init(); err != nil {
		return err
	}

	return nil
}

/**
* GetSeries
* @param kind, tag string
* @return string, error
**/
func GetSeries(kind, tag string) (string, error) {
	if coreSeries == nil {
		return "", fmt.Errorf(MSG_DATABASE_NOT_CONCURRENT)
	}

	item, err := coreSeries.
		Upsert(et.Json{
			"kind": kind,
			"tag":  tag,
		}).
		BeforeInsert(func(tx *Tx, data et.Json) error {
			data.Set("format", "%08d")
			data.Set("value", 1)
			return nil
		}).
		BeforeUpdate(func(tx *Tx, data et.Json) error {
			data.Set("value", ":value + 1")
			return nil
		}).
		Return("value", "format").
		One()
	if err != nil {
		return "", err
	}

	value := item.Int("value")
	format := item.Str("format")
	result := fmt.Sprintf(format, value)

	return result, nil
}

/**
* SetSeries
* @param kind, tag, format string
* @return error
**/
func SetSeries(kind, tag, format string, lastValue int) error {
	if coreSeries == nil {
		return fmt.Errorf(MSG_DATABASE_NOT_CONCURRENT)
	}

	_, err := coreSeries.
		Upsert(et.Json{
			"kind":   kind,
			"tag":    tag,
			"format": format,
			"value":  lastValue,
		}).
		Exec()
	if err != nil {
		return err
	}

	return nil
}

/**
* ImportSeries
* @param items []et.Json
* @return error
**/
func ImportSeries(items []et.Json) error {
	for _, item := range items {
		err := SetSeries(
			item.Str("kind"),
			item.Str("tag"),
			item.Str("format"),
			item.Int("value"),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

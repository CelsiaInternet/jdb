package jdb

import (
	"encoding/json"
	"fmt"

	"github.com/celsiainternet/elvis/et"
)

type Field struct {
	Model      *QlFrom     `json:"model"`
	Name       string      `json:"name"`
	As         string      `json:"as"`
	TypeColumn TypeColumn  `json:"type_column"`
	TypeData   TypeData    `json:"type_data"`
	Default    interface{} `json:"default"`
	Hidden     bool        `json:"hidden"`
}

/**
* Serialize
* @return []byte, error
**/
func (s *Field) Serialize() ([]byte, error) {
	result, err := json.Marshal(s)
	if err != nil {
		return []byte{}, err
	}

	return result, nil
}

/**
* Describe
* @return et.Json
**/
func (s *Field) Describe() et.Json {
	definition, err := s.Serialize()
	if err != nil {
		return et.Json{}
	}

	result := et.Json{}
	err = json.Unmarshal(definition, &result)
	if err != nil {
		return et.Json{}
	}

	return result
}

/**
* asName
* @return string
**/
func (s *Field) asName() string {
	if s.As != "" {
		if s.Model.As != "" {
			return fmt.Sprintf(`%s.%s`, s.Model.As, s.As)
		}
		return s.As
	}

	return s.Name
}

/**
* SetFromAs
* @param as string
* @return *Field
**/
func (s *Field) SetFromAs(as string) *Field {
	s.Model.As = as
	return s
}

/**
* GetField
* @param column *Column
* @return *Field
**/
func GetField(column *Column) *Field {
	return &Field{
		Model: &QlFrom{
			Model: column.model,
			As:    "",
		},
		Name:       column.Name,
		As:         "",
		TypeColumn: column.TypeColumn,
		TypeData:   column.TypeData,
		Default:    column.Default,
		Hidden:     column.Hidden,
	}
}

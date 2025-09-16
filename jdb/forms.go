package jdb

import "github.com/celsiainternet/elvis/strs"

type QlFrom struct {
	*Model
	As string
}

type QlFroms struct {
	Froms []*QlFrom
	index int
}

/**
* newForms
* @return *QlFroms
**/
func newForms() *QlFroms {
	return &QlFroms{
		Froms: make([]*QlFrom, 0),
		index: 65,
	}
}

/**
* add
* @param m *Model
* @return *QlFrom
**/
func (s *QlFroms) add(m *Model) *QlFrom {
	as := string(rune(s.index))
	from := &QlFrom{
		Model: m,
		As:    as,
	}

	s.Froms = append(s.Froms, from)
	s.index++

	return from
}

/**
* getModel
* @param idx int
* @return *Model
**/
func (s *QlFroms) getModel(idx int) *Model {
	if s.Froms[idx] == nil {
		return nil
	}

	return s.Froms[idx].Model
}

/**
* getField
* @param name string
* @return *Field
**/
func (s *QlFroms) getField(name string, create bool) *Field {
	findField := func(name string) *Field {
		for _, from := range s.Froms {
			field := from.getField(name, false)
			if field != nil {
				field.As = from.As
				return field
			}
		}

		return nil
	}

	for tp, ag := range agregations {
		if ag.re.MatchString(name) {
			n := strs.ReplaceAll(name, []string{ag.Agregation, "(", ")"}, "")
			field := findField(n)
			if field != nil {
				field.Agregation = tp
				return field
			}
		}
	}

	return findField(name)
}

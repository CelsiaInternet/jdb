package jdb

import (
	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/reg"
	"github.com/celsiainternet/elvis/strs"
)

/**
* defineColumnIdx
* @param name string, typeData TypeData
* @return *Column
**/
func (s *Model) defineColumnIdx(name string, typeData TypeData, idx int) *Column {
	result := s.getColumn(name)
	if result != nil {
		return result
	}

	result = newColumn(s, name, "", TpColumn, typeData, typeData.DefaultValue())
	if idx == -1 {
		s.addColumn(result)
	} else {
		s.addColumnToIdx(result, idx)
	}

	return result
}

/**
* defineColumn
* @param name string, typeData TypeData
* @return *Column
**/
func (s *Model) defineColumn(name string, typeData TypeData) *Column {
	idx := s.sourceIdx()
	if idx == -1 {
		idx = s.primaryKeyIndex()
	}
	if idx == -1 {
		idx = s.statusIndex()
	}
	if idx == -1 {
		idx = s.projectIndex()
	}
	return s.defineColumnIdx(name, typeData, idx)
}

/**
* defineIndex
* @param sort bool, colums []string
* @return *Model
**/
func (s *Model) defineIndex(sort bool, colums []string) *Model {
	for _, name := range colums {
		col := s.getColumn(name)
		if col == nil {
			continue
		}

		if col.TypeColumn == TpColumn || col.TypeColumn == TpAtribute {
			idx := newIndex(col, sort)
			name := strs.Format("%s_%s_idx", s.Name, col.Name)
			s.Indices[name] = idx
		}
	}

	return s
}

/**
* defineUnique
* @param colums []string
* @return *Model
**/
func (s *Model) defineUnique(colums []string) *Model {
	for _, name := range colums {
		col := s.getColumn(name)
		if col == nil {
			continue
		}

		name := strs.Format("%s_%s_idx", s.Name, col.Name)
		s.Uniques[name] = col
	}

	return s
}

/**
* definePrimaryKey
* @param primaryKeys []string
* @return *Model
**/
func (s *Model) definePrimaryKey(primaryKeys []string) *Model {
	for _, pkn := range primaryKeys {
		col := s.getColumn(pkn)
		if col != nil {
			s.Required[col.Name] = true
			s.PrimaryKeys[col.Name] = col
		}
	}

	return s
}

/**
* definePrimaryKeyField
* @return *Column
**/
func (s *Model) definePrimaryKeyField() *Column {
	result := s.defineColumn(cf.Key, TypeDataKey)
	s.definePrimaryKey([]string{cf.Key})

	return result
}

/**
* defineForeignKey
* @param fks map[string]string, withName string, onDeleteCascade, onUpdateCascade bool
* @return *Relation
**/
func (s *Model) defineForeignKey(fks map[string]string, withName string, onDeleteCascade, onUpdateCascade bool) *Relation {
	with := s.GetModel(withName)
	if with == nil {
		return nil
	}

	result := &Relation{
		With:            with,
		Fk:              fks,
		Limit:           0,
		OnDeleteCascade: onDeleteCascade,
		OnUpdateCascade: onUpdateCascade,
	}
	name := strs.Format("%s_%s_fk", s.Name, with.Name)
	s.ForeignKeys[name] = result
	s.RelationsTo[withName] = result

	return result
}

/**
* defineSource
* @param name string
* @return *Column
**/
func (s *Model) defineSource(name string) *Column {
	if s.SourceField != nil {
		return s.SourceField
	}

	result := s.defineColumn(name, TypeDataObject)
	s.defineIndex(true, []string{name})
	s.SourceField = result

	return result
}

/**
* defineSourceField
* @return *Column
**/
func (s *Model) defineSourceField() *Column {
	if s.SourceField != nil {
		return s.SourceField
	}

	return s.defineSource(cf.Source)
}

/**
* defineIndexField
* @return *Column
**/
func (s *Model) defineIndexField() *Column {
	if s.IndexField != nil {
		return s.IndexField
	}

	result := s.defineColumn(cf.Index, TypeDataText)
	result.Hidden = true
	s.BeforeDelete(func(tx *Tx, data et.Json) error {
		data.Set(cf.Index, reg.ULID())
		return nil
	})
	s.IndexField = result

	return result
}

/**
* defineAtribute
* @param name string, typeData TypeData
* @return *Column
**/
func (s *Model) defineAtribute(name string, typeData TypeData) *Column {
	s.defineSourceField()
	result := newAtribute(s, name, typeData)
	s.addColumn(result)

	return result
}

/**
* defineCreatedAtField
* @return *Column
**/
func (s *Model) defineCreatedAtField() *Column {
	result := s.defineColumn(cf.CreatedAt, TypeDataDateTime)
	s.defineIndex(true, []string{cf.CreatedAt})
	s.CreatedAtField = result

	return result
}

/**
* defineUpdatedAtField
* @return *Column
**/
func (s *Model) defineUpdatedAtField() *Column {
	result := s.defineColumn(cf.UpdatedAt, TypeDataDateTime)
	s.defineIndex(true, []string{cf.UpdatedAt})
	s.UpdatedAtField = result

	return result
}

/**
* defineStatusField
* @return *Column
**/
func (s *Model) defineStatusField() *Column {
	result := s.defineColumn(cf.StatusId, TypeDataState)
	s.defineIndex(true, []string{cf.StatusId})
	s.StatusField = result

	return result
}

/**
* defineSystemKeyField
* @return *Column
**/
func (s *Model) defineSystemKeyField() *Column {
	result := s.defineColumn(cf.SystemId, TypeDataKey)
	result.Hidden = true
	s.defineIndex(true, []string{cf.SystemId})
	s.SystemKeyField = result

	return result
}

/**
* defineProjectField
* @return *Column
**/
func (s *Model) defineProjectField() *Column {
	result := s.defineColumn(cf.ProjectId, TypeDataKey)
	s.defineIndex(true, []string{cf.ProjectId})
	s.ProjectField = result

	return result
}

/**
* defineHidden
* @param colums []string
* @return *Model
**/
func (s *Model) defineHidden(colums []string) *Model {
	for _, name := range colums {
		col := s.getColumn(name)
		if col != nil {
			col.Hidden = true
		}
	}

	return s
}

/**
* defineRequired
* @param colums []string
* @return *Model
**/
func (s *Model) defineRequired(colums []string) *Model {
	for _, name := range colums {
		col := s.getColumn(name)
		if col != nil {
			s.Required[name] = true
		}
	}

	return s
}

/**
* defineRelation
* @param name, withName string, fks map[string]string, limit int
* @return *Relation
**/
func (s *Model) defineRelation(name, withName string, fks map[string]string, limit int) *Relation {
	with := s.GetModel(withName)
	if with == nil {
		return nil
	}

	col := newColumn(s, name, "", TpRelatedTo, TypeDataNone, TypeDataNone.DefaultValue())
	s.addColumn(col)
	result := s.defineForeignKey(fks, withName, true, true)
	s.RelationsTo[name] = result
	return result
}

/**
* defineRollup
* @param name, rollupFrom string, fks map[string]string, fields []string, showRollup ShowRollup
* @return *Rollup
**/
func (s *Model) defineRollup(name, withName string, fks map[string]string, fields []string, showRollup ShowRollup) *Rollup {
	with := s.GetModel(withName)
	if with == nil {
		return nil
	}

	result := &Rollup{
		With:   with,
		Fk:     fks,
		Fields: fields,
		Show:   showRollup,
	}

	col := newColumn(s, name, "", TpRollup, TypeDataNone, TypeDataNone.DefaultValue())
	s.addColumn(col)
	s.Rollup[name] = result
	return result
}

/**
* defineDetail
* @param name string, fks map[string]string, limit int
* @return *Model
**/
func (s *Model) defineDetail(name string, fks map[string]string, limit int) *Model {
	withName := s.Name + "_" + name
	with := s.GetModel(withName)
	if with == nil {
		with = NewModel(s.schema, withName, 1)
		with.definePrimaryKeyField()
		for fkn := range fks {
			col := with.defineColumn(fkn, TypeDataKey)
			with.definePrimaryKey([]string{col.Name})
		}
	}

	col := newColumn(s, name, "", TpDetail, TypeDataNone, TypeDataNone.DefaultValue())
	s.addColumn(col)
	result := s.defineForeignKey(fks, withName, true, true)
	s.Detail[name] = result
	return with
}

/**
* defineModel
* @return *Model
**/
func (s *Model) defineModel() *Model {
	s.defineCreatedAtField()
	s.defineUpdatedAtField()
	s.defineStatusField()
	s.definePrimaryKeyField()
	s.defineSourceField()
	s.defineSystemKeyField()
	s.defineIndexField()

	return s
}

/**
* defineProjectModel
* @return *Model
**/
func (s *Model) defineProjectModel() *Model {
	s.defineCreatedAtField()
	s.defineUpdatedAtField()
	s.defineProjectField()
	s.defineStatusField()
	s.definePrimaryKeyField()
	s.defineSourceField()
	s.defineSystemKeyField()
	s.defineIndexField()

	return s
}

/**
* DefineIntegrity
**/
func (s *Model) DefineIntegrity() *Model {
	s.Integrity = true
	return s
}

/**
* DefineColumn
* @param name string, typeData TypeData
* @return *Column
**/
func (s *Model) DefineColumn(name string, typeData TypeData) *Column {
	return s.defineColumn(name, typeData)
}

/**
* DefineIndex
* @param sort bool, colums []string
* @return *Model
**/
func (s *Model) DefineIndex(sort bool, colums ...string) *Model {
	return s.defineIndex(sort, colums)
}

/**
* DefineUnique
* @param colums ...string
* @return *Model
**/
func (s *Model) DefineUnique(colums ...string) *Model {
	return s.defineUnique(colums)
}

/**
* DefinePrimaryKey
* @param colums ...string
* @return *Model
**/
func (s *Model) DefinePrimaryKey(colums ...string) *Model {
	return s.definePrimaryKey(colums)
}

/**
* DefinePrimaryKeyField
* @return *Model
**/
func (s *Model) DefinePrimaryKeyField() *Model {
	s.definePrimaryKeyField()
	return s
}

/**
* DefineForeignKey
* @param fks map[string]string, withName string, onDeleteCascade, onUpdateCascade bool
* @return *Model
**/
func (s *Model) DefineForeignKey(fks map[string]string, withName string, onDeleteCascade, onUpdateCascade bool) *Model {
	s.defineForeignKey(fks, withName, onDeleteCascade, onUpdateCascade)
	return s
}

/**
* DefineSource
* @param name string
* @return *Column
**/
func (s *Model) DefineSource(name string) *Column {
	return s.defineSource(name)
}

/**
* DefineSourceField
* @return *Column
**/
func (s *Model) DefineSourceField() *Column {
	return s.defineSourceField()
}

/**
* DefineIndexField
* @return *Column
**/
func (s *Model) DefineIndexField() *Column {
	return s.defineIndexField()
}

/**
* DefineAtribute
* @param name string, typeData TypeData
* @return *Column
**/
func (s *Model) DefineAtribute(name string, typeData TypeData) *Column {
	return s.defineAtribute(name, typeData)
}

/**
* DefineCreatedAtField
* @return *Column
**/
func (s *Model) DefineCreatedAtField() *Column {
	return s.defineCreatedAtField()
}

/**
* DefineUpdatedAtField
* @return *Column
**/
func (s *Model) DefineUpdatedAtField() *Column {
	return s.defineUpdatedAtField()
}

/**
* DefineStatusField
* @return *Column
**/
func (s *Model) DefineStatusField() *Column {
	return s.defineStatusField()
}

/**
* DefineSystemKeyField
* @return *Column
**/
func (s *Model) DefineSystemKeyField() *Column {
	return s.defineSystemKeyField()
}

/**
* DefineProjectField
* @return *Column
**/
func (s *Model) DefineProjectField() *Column {
	return s.defineProjectField()
}

/**
* DefineHidden
* @param colums ...string
* @return *Model
**/
func (s *Model) DefineHidden(colums ...string) *Model {
	s.defineHidden(colums)
	return s
}

/**
* DefineRequired
* @param colums ...string
* @return *Model
**/
func (s *Model) DefineRequired(colums ...string) *Model {
	s.defineRequired(colums)
	return s
}

/**
* DefineRelation
* @param name, withName string, fks map[string]string, limit int
* @return *Model
**/
func (s *Model) DefineRelation(name, withName string, fks map[string]string, limit int) *Relation {
	return s.defineRelation(name, withName, fks, limit)
}

/**
* DefineRollup
* @param name, withName string, fks map[string]string, fields []string
* @return *Model
**/
func (s *Model) DefineRollup(name, withName string, fks map[string]string, fields []string) *Rollup {
	return s.defineRollup(name, withName, fks, fields, ShowAtrib)
}

/**
* DefineObject
* @param name, rollupFrom string, fks map[string]string, field string
* @return *Model
**/
func (s *Model) DefineObject(name, withName string, fks map[string]string, fields []string) *Model {
	s.defineRollup(name, withName, fks, fields, ShowObject)
	return s
}

/**
* DefineDetail
* @param name string, fks map[string]string, limit int
* @return *Model
**/
func (s *Model) DefineDetail(name string, fks map[string]string, limit int) *Model {
	return s.defineDetail(name, fks, limit)
}

/**
* DefineModel
* @return *Model
**/
func (s *Model) DefineModel() *Model {
	return s.defineModel()
}

/**
* DefineProjectModel
* @return *Model
**/
func (s *Model) DefineProjectModel() *Model {
	return s.defineProjectModel()
}

/**
* DefineCalc
* @param name string, fn DataFunctionTx
* @return Model
**/
func (s *Model) DefineCalc(name string, fn DataFunctionTx) *Model {
	result := s.getColumn(name)
	if result != nil {
		return s
	}

	result = newColumn(s, name, "", TpCalc, TypeDataNone, TypeDataNone.DefaultValue())
	s.addColumn(result)
	s.CalcFunction[name] = fn
	return s
}

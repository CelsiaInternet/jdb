package jdb

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/event"
	"github.com/celsiainternet/elvis/strs"
	"github.com/celsiainternet/elvis/timezone"
	"github.com/google/uuid"
)

type TypeId int

const (
	TpNodeId TypeId = iota
	TpUUId
	TpULId
	TpXId
)

func (s TypeId) Str() string {
	switch s {
	case TpUUId:
		return "uuid"
	case TpULId:
		return "ulid"
	case TpXId:
		return "xid"
	default:
		return "id"
	}
}

var (
	ErrNotInserted = fmt.Errorf("record not inserted")
	ErrNotUpdated  = fmt.Errorf("record not updated")
	ErrNotFound    = fmt.Errorf("record not found")
	ErrNotUpserted = fmt.Errorf("record not inserted or updated")
	ErrDuplicate   = fmt.Errorf("record duplicate")
)

type Model struct {
	Db                  *DB                       `json:"-"`
	schema              *Schema                   `json:"-"`
	Schema              string                    `json:"schema"`
	Table               string                    `json:"table"`
	CreatedAt           time.Time                 `json:"created_at"`
	UpdateAt            time.Time                 `json:"updated_at"`
	Id                  string                    `json:"id"`
	Name                string                    `json:"name"`
	Description         string                    `json:"description"`
	UseCore             bool                      `json:"use_core"`
	Integrity           bool                      `json:"integrity"`
	Columns             []*Column                 `json:"-"`
	PrimaryKeys         map[string]*Column        `json:"-"`
	ForeignKeys         map[string]*Relation      `json:"-"`
	Indices             map[string]*Index         `json:"-"`
	Uniques             map[string]*Column        `json:"-"`
	Required            map[string]bool           `json:"-"`
	Detail              map[string]*Relation      `json:"detail"`
	Rollup              map[string]*Rollup        `json:"rollup"`
	FullText            map[string]*FullText      `json:"fulltext"`
	CalcFunction        map[string]DataFunctionTx `json:"-"`
	RelationsTo         map[string]*Relation      `json:"-"`
	CreatedAtField      *Column                   `json:"-"`
	UpdatedAtField      *Column                   `json:"-"`
	SystemKeyField      *Column                   `json:"-"`
	StatusField         *Column                   `json:"-"`
	IndexField          *Column                   `json:"-"`
	SourceField         *Column                   `json:"-"`
	FullTextField       *Column                   `json:"-"`
	ProjectField        *Column                   `json:"-"`
	Version             int                       `json:"version"`
	beforeInsert        []DataFunctionTx          `json:"-"`
	beforeUpdate        []DataFunctionTx          `json:"-"`
	beforeDelete        []DataFunctionTx          `json:"-"`
	afterInsert         []DataFunctionTx          `json:"-"`
	afterUpdate         []DataFunctionTx          `json:"-"`
	afterDelete         []DataFunctionTx          `json:"-"`
	beforeInsertTrigger []TriggerFunctionTx       `json:"-"`
	beforeUpdateTrigger []TriggerFunctionTx       `json:"-"`
	beforeDeleteTrigger []TriggerFunctionTx       `json:"-"`
	afterInsertTrigger  []TriggerFunctionTx       `json:"-"`
	afterUpdateTrigger  []TriggerFunctionTx       `json:"-"`
	afterDeleteTrigger  []TriggerFunctionTx       `json:"-"`
	eventEmiterChannel  chan event.EvenMessage    `json:"-"`
	eventsEmiter        map[string]event.Handler  `json:"-"`
	IsDebug             bool                      `json:"-"`
	isLocked            bool                      `json:"-"`
	isInit              bool                      `json:"-"`
	needMutate          bool                      `json:"-"`
}

/**
* NewTable
* @param db *DB, table string
* @return *Model
**/
func NewTable(db *DB, table string) *Model {
	idx := slices.IndexFunc(db.tables, func(e *Model) bool { return e.Name == table })
	if idx != -1 {
		return db.tables[idx]
	}

	list := strs.Split(table, ".")
	if len(list) < 2 {
		return nil
	}
	name := list[1]
	schemaName := list[0]

	schema := NewSchema(db, schemaName)
	now := timezone.NowTime()
	result := &Model{
		Db:                  db,
		schema:              schema,
		Schema:              schema.Name,
		Table:               name,
		CreatedAt:           now,
		UpdateAt:            now,
		Id:                  fmt.Sprintf("%s.%s.%s", schema.Db.Name, schema.Name, name),
		Name:                table,
		UseCore:             false,
		Columns:             make([]*Column, 0),
		PrimaryKeys:         make(map[string]*Column),
		ForeignKeys:         make(map[string]*Relation),
		Indices:             make(map[string]*Index),
		Uniques:             make(map[string]*Column),
		Required:            make(map[string]bool),
		Detail:              make(map[string]*Relation),
		Rollup:              make(map[string]*Rollup),
		FullText:            make(map[string]*FullText),
		CalcFunction:        make(map[string]DataFunctionTx),
		RelationsTo:         make(map[string]*Relation),
		beforeInsert:        []DataFunctionTx{},
		beforeUpdate:        []DataFunctionTx{},
		beforeDelete:        []DataFunctionTx{},
		afterInsert:         []DataFunctionTx{},
		afterUpdate:         []DataFunctionTx{},
		afterDelete:         []DataFunctionTx{},
		beforeInsertTrigger: []TriggerFunctionTx{},
		beforeUpdateTrigger: []TriggerFunctionTx{},
		beforeDeleteTrigger: []TriggerFunctionTx{},
		afterInsertTrigger:  []TriggerFunctionTx{},
		afterUpdateTrigger:  []TriggerFunctionTx{},
		afterDeleteTrigger:  []TriggerFunctionTx{},
		eventEmiterChannel:  make(chan event.EvenMessage),
		eventsEmiter:        make(map[string]event.Handler),
		Version:             1,
		IsDebug:             db.IsDebug,
	}
	result.AfterInsert(result.afterInsertDefault)
	result.AfterUpdate(result.afterUpdateDefault)
	result.AfterDelete(result.afterDeleteDefault)
	db.tables = append(db.tables, result)

	return result
}

/**
* NewModel
* @param schema *Schema, name string, version int
* @return *Model
**/
func NewModel(schema *Schema, name string, version int) *Model {
	idx := slices.IndexFunc(schema.Db.models, func(e *Model) bool { return e.Name == name })
	if idx != -1 {
		return schema.Db.models[idx]
	}

	now := timezone.NowTime()
	result := &Model{
		Db:                  schema.Db,
		schema:              schema,
		Schema:              schema.Name,
		Table:               name,
		CreatedAt:           now,
		UpdateAt:            now,
		Id:                  fmt.Sprintf("%s.%s.%s", schema.Db.Name, schema.Name, name),
		Name:                name,
		UseCore:             schema.UseCore,
		Columns:             make([]*Column, 0),
		PrimaryKeys:         make(map[string]*Column),
		ForeignKeys:         make(map[string]*Relation),
		Indices:             make(map[string]*Index),
		Uniques:             make(map[string]*Column),
		Required:            make(map[string]bool),
		Detail:              make(map[string]*Relation),
		Rollup:              make(map[string]*Rollup),
		FullText:            make(map[string]*FullText),
		CalcFunction:        make(map[string]DataFunctionTx),
		RelationsTo:         make(map[string]*Relation),
		beforeInsert:        []DataFunctionTx{},
		beforeUpdate:        []DataFunctionTx{},
		beforeDelete:        []DataFunctionTx{},
		afterInsert:         []DataFunctionTx{},
		afterUpdate:         []DataFunctionTx{},
		afterDelete:         []DataFunctionTx{},
		beforeInsertTrigger: []TriggerFunctionTx{},
		beforeUpdateTrigger: []TriggerFunctionTx{},
		beforeDeleteTrigger: []TriggerFunctionTx{},
		afterInsertTrigger:  []TriggerFunctionTx{},
		afterUpdateTrigger:  []TriggerFunctionTx{},
		afterDeleteTrigger:  []TriggerFunctionTx{},
		eventEmiterChannel:  make(chan event.EvenMessage),
		eventsEmiter:        make(map[string]event.Handler),
		Version:             version,
		IsDebug:             schema.Db.IsDebug,
	}
	result.AfterInsert(result.afterInsertDefault)
	result.AfterUpdate(result.afterUpdateDefault)
	result.AfterDelete(result.afterDeleteDefault)

	schema.addModel(result)
	return result
}

/**
* GetModel
* @param name string
* @return *Model
**/
func (s *Model) GetModel(name string) *Model {
	idx := slices.IndexFunc(s.Db.models, func(e *Model) bool { return e.Name == name })
	if idx != -1 {
		return s.Db.models[idx]
	}

	return NewModel(s.schema, name, 1)
}

/**
* Serialize
* @return []byte, error
**/
func (s *Model) serialize() ([]byte, error) {
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
func (s *Model) Describe() et.Json {
	definition, err := s.serialize()
	if err != nil {
		return et.Json{}
	}

	result := et.Json{}
	err = json.Unmarshal(definition, &result)
	if err != nil {
		return et.Json{}
	}

	columns := make([]et.Json, 0)
	for _, column := range s.Columns {
		columns = append(columns, column.Describe())
	}

	delete(result, "definitions")
	result["kind"] = "model"
	result["columns"] = columns
	result["primary_keys"] = s.PrimaryKeys
	result["foreign_keys"] = s.ForeignKeys
	result["indices"] = s.Indices
	result["uniques"] = s.Uniques
	result["relations_to"] = s.RelationsTo
	result["required"] = s.Required
	result["system_key_field"] = s.SystemKeyField
	result["status_field"] = s.StatusField
	result["index_field"] = s.IndexField
	result["source_field"] = s.SourceField
	result["full_text_field"] = s.FullTextField
	result["project_field"] = s.ProjectField

	return result
}

/**
* Save
* @return error
**/
func (s *Model) Save() error {
	if s == nil || !s.UseCore || !s.Db.isInit {
		return nil
	}

	definition, err := s.serialize()
	if err != nil {
		return err
	}

	err = s.Db.upsertModel("model", s.Name, s.Version, definition)
	if err != nil {
		return err
	}

	s.isInit = true

	return nil
}

/**
* Drop
**/
func (s *Model) Drop() {
	if s.Db == nil {
		return
	}

	for _, detail := range s.RelationsTo {
		model := detail.With
		if model != nil && model.Name != s.Name {
			model.Drop()
		}
	}

	s.Db.DropModel(s)
}

/**
* Empty
**/
func (s *Model) Empty() {
	if s.Db == nil {
		return
	}

	for _, detail := range s.RelationsTo {
		model := detail.With
		if model != nil && model.Name != s.Name {
			model.Empty()
		}
	}

	s.Db.EmptyModel(s)
}

/**
* Init
* @return error
**/
func (s *Model) Init() error {
	if s.isInit {
		return nil
	}

	if s.SourceField != nil {
		idx := s.SourceField.idx()
		if idx != len(s.Columns)-1 && idx != -1 {
			s.moveColumnToEnd(s.SourceField, idx)
		}
	}

	if s.SystemKeyField != nil {
		idx := s.SystemKeyField.idx()
		if idx != len(s.Columns)-1 && idx != -1 {
			s.moveColumnToEnd(s.SystemKeyField, idx)
		}
	}

	if s.IndexField != nil {
		idx := s.IndexField.idx()
		if idx != len(s.Columns)-1 && idx != -1 {
			s.moveColumnToEnd(s.IndexField, idx)
		}
	}

	go func() {
		for message := range s.eventEmiterChannel {
			s.eventEmiter(message)
		}
	}()

	exist, err := s.Db.LoadModel(s)
	if err != nil {
		return err
	}

	s.isInit = true

	if exist {
		return nil
	}

	err = s.Save()
	if err != nil {
		return err
	}

	return nil
}

/**
* CheckRequired
* @param data et.Json
* @return error
**/
func (s *Model) CheckRequired(data et.Json) error {
	for name, required := range s.Required {
		if required {
			if data[name] == nil {
				return fmt.Errorf(MSG_REQUIRED_FIELD_REQUIRED, name)
			}
		}
	}

	return nil
}

/**
* GetId
* @param id string
* @return string
**/
func (s *Model) GetId(id string) string {
	if !map[string]bool{"": true, "*": true, "new": true}[id] {
		return id
	}
	return strs.Format(`%s`, uuid.NewString())
}

/**
* GenId
* @return string
**/
func (s *Model) GenId() string {
	return s.GetId("new")
}

/**
* sourceIdx
* @return int
**/
func (s *Model) sourceIdx() int {
	if s.SourceField == nil {
		return -1
	}

	return s.SourceField.idx()
}

/**
* primaryKeyIndex
* @return int
**/
func (s *Model) primaryKeyIndex() int {
	result := -1
	for _, pk := range s.PrimaryKeys {
		if pk.idx() > result {
			result = pk.idx()
		}
	}

	return result
}

/**
* statusIndex
* @return int
**/
func (s *Model) statusIndex() int {
	if s.StatusField == nil {
		return -1
	}

	return s.StatusField.idx()
}

/**
* projectIndex
* @return int
**/
func (s *Model) projectIndex() int {
	if s.ProjectField == nil {
		return -1
	}

	return s.ProjectField.idx()
}

/**
* Debug
* @return *Model
**/
func (s *Model) Debug() *Model {
	s.IsDebug = true

	return s
}

/**
* addColumn
* @param column *Column
**/
func (s *Model) addColumn(column *Column) {
	idx := slices.IndexFunc(s.Columns, func(e *Column) bool { return strings.ToLower(e.Name) == strings.ToLower(column.Name) })
	if idx == -1 {
		s.Columns = append(s.Columns, column)
	}
}

/**
* addColumnIdx
* @param column *Column, idx int
**/
func (s *Model) addColumnToIdx(column *Column, idx int) {
	if idx != -1 {
		s.Columns = append(s.Columns[:idx], append([]*Column{column}, s.Columns[idx:]...)...)
	}
}

/**
* moveColumnToEnd
* @param column *Column, idx int
**/
func (s *Model) moveColumnToEnd(column *Column, idx int) {
	s.Columns = append(s.Columns[:idx], s.Columns[idx+1:]...)
	s.addColumn(column)
}

/**
* getColumn
* @param name string
* @return *Column
**/
func (s *Model) getColumn(name string) *Column {
	for _, col := range s.Columns {
		if col.Name == name {
			return col
		}
	}

	return nil
}

/**
* GetField
* @param name string, isCreate bool
* @return *Field
**/
func (s *Model) getField(name string, isCreate bool) *Field {
	getField := func(name string) *Field {
		col := s.getColumn(name)
		if col != nil {
			return GetField(col)
		}

		if !isCreate {
			return nil
		}

		if s.Integrity {
			return nil
		}

		if s.SourceField == nil {
			return nil
		}

		result := newAtribute(s, name, TypeDataText)
		return GetField(result)
	}

	split := strings.Split(name, ":")
	as := name
	if len(split) > 1 {
		name = split[0]
		as = split[1]
	}

	result := getField(name)
	if result != nil {
		result.As = as
		return result
	}

	return nil
}

/**
* GetField
* @param name string
* @return *Field
**/
func (s *Model) GetField(name string) *Field {
	return s.getField(name, false)
}

/**
* Counted
* @return int, error
**/
func (s *Model) CountedTx(tx *Tx) (int, error) {
	all, err := From(s).
		CountedTx(tx)
	if err != nil {
		return 0, err
	}

	return all, nil
}

/**
* Counted
* @return int, error
**/
func (s *Model) Counted() (int, error) {
	return s.CountedTx(nil)
}

/**
* QueryTx
* @param tx *Tx, params et.Json
* @return et.Json, error
**/
func (s *Model) QueryTx(tx *Tx, query et.Json) (et.Json, error) {
	return From(s).
		queryTx(tx, query)
}

/**
* Query
* @param params et.Json
* @return et.Json, error
**/
func (s *Model) Query(params et.Json) (et.Json, error) {
	return s.QueryTx(nil, params)
}

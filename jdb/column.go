package jdb

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/timezone"
)

type TypeColumn int

const (
	TpColumn TypeColumn = iota
	TpAtribute
	TpRollup
	TpDetail
	TpRelatedTo
	TpCalc
)

func (s TypeColumn) Str() string {
	switch s {
	case TpColumn:
		return "column"
	case TpRollup:
		return "rollup"
	case TpRelatedTo:
		return "related_to"
	case TpCalc:
		return "calc"
	default:
		return "attribute"
	}
}

/**
* StrsToTypeColumn
* @param strs string
* @return TypeColumn
**/
func StrsToTypeColumn(strs string) TypeColumn {
	switch strs {
	case "column":
		return TpColumn
	case "attribute":
		return TpAtribute
	case "calc":
		return TpCalc
	case "related_to":
		return TpRelatedTo
	case "rollup":
		return TpRollup
	}

	return TpAtribute
}

type TypeData int

const (
	TypeDataText TypeData = iota
	TypeDataMemo
	TypeDataShortText
	TypeDataKey
	TypeDataNumber
	TypeDataInt
	TypeDataPrecision
	TypeDataDateTime
	TypeDataCheckbox
	TypeDataBytes
	/* Jsonb */
	TypeDataObject
	TypeDataSelect
	TypeDataMultiSelect
	TypeDataGeometry
	/* Fulltext */
	TypeDataFullText
	/* Objects */
	TypeDataState
	TypeDataUser
	TypeDataFilesMedia
	TypeDataUrl
	TypeDataEmail
	TypeDataPhone
	TypeDataAddress
	TypeDataRelation
	TypeDataRollup
	/* None */
	TypeDataNone
	// Status
	Active   = "0"
	Archive  = "1"
	Cancel   = "2"
	System   = "-1"
	ToDelete = "-2"
)

func (s TypeData) DefaultValue() interface{} {
	switch s {
	case TypeDataNumber:
		return 0.0
	case TypeDataInt:
		return 0
	case TypeDataCheckbox:
		return false
	case TypeDataSelect:
		return ""
	case TypeDataMultiSelect:
		return []string{}
	case TypeDataMemo:
		return ""
	case TypeDataObject:
		return et.Json{}
	case TypeDataGeometry:
		return et.Json{
			"type":        "Point",
			"coordinates": []float64{0.00, 0.00},
		}
	}

	return ""
}

func (s TypeData) Str() string {
	switch s {
	case TypeDataText:
		return "text"
	case TypeDataMemo:
		return "memo"
	case TypeDataShortText:
		return "short-text"
	case TypeDataKey:
		return "key"
	case TypeDataNumber:
		return "number"
	case TypeDataInt:
		return "int"
	case TypeDataPrecision:
		return "precision"
	case TypeDataDateTime:
		return "date-time"
	case TypeDataBytes:
		return "bytes"
	case TypeDataObject:
		return "object"
	case TypeDataFullText:
		return "full-text"
	case TypeDataSelect:
		return "select"
	case TypeDataMultiSelect:
		return "multi-select"
	case TypeDataGeometry:
		return "location"
	case TypeDataCheckbox:
		return "checkbox"
	case TypeDataState:
		return "status"
	case TypeDataUser:
		return "person"
	case TypeDataFilesMedia:
		return "files-media"
	case TypeDataUrl:
		return "url"
	case TypeDataEmail:
		return "email"
	case TypeDataPhone:
		return "phone"
	case TypeDataAddress:
		return "address"
	case TypeDataRelation:
		return "relation"
	case TypeDataRollup:
		return "rollup"
	default:
		return "none"
	}
}

func StrsToTypeData(strs string) TypeData {
	switch strs {
	case "text":
		return TypeDataText
	case "memo":
		return TypeDataMemo
	case "number":
		return TypeDataNumber
	case "select":
		return TypeDataSelect
	case "multi-select":
		return TypeDataMultiSelect
	case "date-time":
		return TypeDataDateTime
	case "checkbox":
		return TypeDataCheckbox
	case "status":
		return TypeDataState
	case "person":
		return TypeDataUser
	case "files-media":
		return TypeDataFilesMedia
	case "url":
		return TypeDataUrl
	case "email":
		return TypeDataEmail
	case "phone":
		return TypeDataPhone
	case "address":
		return TypeDataAddress
	case "location":
		return TypeDataGeometry
	case "relation":
		return TypeDataRelation
	case "rollup":
		return TypeDataRollup
	}

	return TypeDataText
}

func StrToKindType(strs string) (TypeColumn, TypeData) {
	switch strs {
	case "text":
		return TpAtribute, TypeDataText
	case "memo":
		return TpAtribute, TypeDataMemo
	case "number":
		return TpAtribute, TypeDataNumber
	case "select":
		return TpAtribute, TypeDataSelect
	case "multi-select":
		return TpAtribute, TypeDataMultiSelect
	case "date-time":
		return TpAtribute, TypeDataDateTime
	case "checkbox":
		return TpAtribute, TypeDataCheckbox
	case "status":
		return TpColumn, TypeDataState
	case "person":
		return TpColumn, TypeDataUser
	case "files-media":
		return TpColumn, TypeDataFilesMedia
	case "url":
		return TpColumn, TypeDataUrl
	case "email":
		return TpColumn, TypeDataEmail
	case "phone":
		return TpColumn, TypeDataPhone
	case "address":
		return TpAtribute, TypeDataAddress
	case "location":
		return TpColumn, TypeDataGeometry
	case "relation":
		return TpRelatedTo, TypeDataRelation
	case "rollup":
		return TpRollup, TypeDataRollup
	}

	return TpAtribute, TypeDataText
}

type ColumnFields struct {
	Key       string
	Source    string
	Index     string
	ProjectId string
	CreatedAt string
	UpdatedAt string
	StatusId  string
	SystemId  string
	CreatedTo string
	UpdatedTo string
	Fulltext  string
	Checked   string
}

/**
* Json
* @return et.Json
**/
func (s *ColumnFields) Json() et.Json {
	return et.Json{
		"key":        s.Key,
		"source":     s.Source,
		"index":      s.Index,
		"project_id": s.ProjectId,
		"created_at": s.CreatedAt,
		"updated_at": s.UpdatedAt,
		"status_id":  s.StatusId,
		"system_id":  s.SystemId,
		"created_to": s.CreatedTo,
		"updated_to": s.UpdatedTo,
		"fulltext":   s.Fulltext,
		"checked":    s.Checked,
	}
}

var (
	cf         *ColumnFields
	KEY        string
	INDEX      string
	SOURCE     string
	PROJECT_ID string
	CREATED_AT string
	UPDATED_AT string
	STATUS_ID  string
	SYSID      string
	CREATED_TO string
	UPDATED_TO string
	FULLTEXT   string
	HISTORYCAL string
	CHECKED    string
)

func init() {
	cf = &ColumnFields{
		Key:       "_id",
		Source:    "_data",
		Index:     "idx",
		ProjectId: "project_id",
		CreatedAt: "created_at",
		UpdatedAt: "updated_at",
		StatusId:  "_status",
		SystemId:  "jdbid",
		CreatedTo: "created_to",
		UpdatedTo: "updated_to",
		Fulltext:  "fulltext",
		Checked:   "checked",
	}

	SetColumnFields(cf)
}

/**
* SetColumnFields
* @param fields *ColumnFields
**/
func SetColumnFields(fields *ColumnFields) {
	cf = fields
	KEY = cf.Key
	SOURCE = cf.Source
	PROJECT_ID = cf.ProjectId
	CREATED_AT = cf.CreatedAt
	UPDATED_AT = cf.UpdatedAt
	STATUS_ID = cf.StatusId
	SYSID = cf.SystemId
	CREATED_TO = cf.CreatedTo
	UPDATED_TO = cf.UpdatedTo
	FULLTEXT = cf.Fulltext
	CHECKED = cf.Checked
}

type Relation struct {
	With            *Model            `json:"-"`
	Fk              map[string]string `json:"fk"`
	Limit           int               `json:"rows"`
	OnDeleteCascade bool              `json:"on_delete_cascade"`
	OnUpdateCascade bool              `json:"on_update_cascade"`
	IsMultiSelect   bool              `json:"is_multi_select"`
}

/**
* GetWhere
* @param from et.Json
* @return et.Json
**/
func (s *Relation) GetWhere(from et.Json) et.Json {
	result := et.Json{}
	for fkn, pkn := range s.Fk {
		fk := from.Get(fkn)
		if fk == nil {
			continue
		}

		result[pkn] = et.Json{
			"eq": fk,
		}
	}

	return result
}

type ShowRollup int

const (
	ShowAtrib ShowRollup = iota
	ShowObject
)

/**
* Str
* @return string
**/
func (s ShowRollup) Str() string {
	switch s {
	case ShowObject:
		return "objects"
	default:
		return "attribs"
	}
}

/**
* MarshalJSON: Serializa el modo como texto ("attribs" u "objects").
* @return []byte, error
**/
func (s ShowRollup) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Str())
}

/**
* UnmarshalJSON: Lee el modo como texto ("attribs" u "objects") o como numero (definiciones anteriores).
* @param data []byte
* @return error
**/
func (s *ShowRollup) UnmarshalJSON(data []byte) error {
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		*s = ShowAtrib
		if text == ShowObject.Str() {
			*s = ShowObject
		}
		return nil
	}

	var number int
	if err := json.Unmarshal(data, &number); err != nil {
		return err
	}

	*s = ShowRollup(number)
	return nil
}

/**
* Rollup: Campos de otro modelo (With) que se agregan a cada fila, buscados por las llaves Fk (campo local -> campo en With).
* Name es la columna virtual del modelo y WithName el nombre de With, para reconstruirlo desde la definicion serializada.
* Wheres son condiciones fijas sobre With ({ campo: { operador: valor } }), unidas con AND a las llaves.
**/
type Rollup struct {
	Name     string            `json:"name"`
	With     *Model            `json:"-"`
	WithName string            `json:"with"`
	Fk       map[string]string `json:"fk"`
	Fields   []string          `json:"fields"`
	Show     ShowRollup        `json:"show"`
	Wheres   []et.Json         `json:"wheres"`
}

// Operadores que acepta una condicion de Rollup.Wheres (los de QlWhere.setWheres).
var rollupOperators = map[string]bool{
	"eq": true, "neg": true, "in": true, "notIn": true, "like": true, "more": true, "less": true,
	"moreEq": true, "lessEq": true, "between": true, "isNull": true, "notNull": true,
}

/**
* SetWheres: Condiciones fijas sobre With, unidas con AND a las llaves; reemplaza las anteriores.
* Cada una es { campo: { operador: valor } }, ej. et.Json{"key": et.Json{"eq": "CC"}}.
* @param wheres ...et.Json
* @return *Rollup, error
**/
func (s *Rollup) SetWheres(wheres ...et.Json) (*Rollup, error) {
	for _, where := range wheres {
		for field, condition := range where {
			operators, ok := condition.(map[string]interface{})
			if jsonCondition, isJson := condition.(et.Json); isJson {
				operators, ok = jsonCondition, true
			}
			if !ok || len(operators) == 0 {
				return s, fmt.Errorf(MSG_ROLLUP_WHERE_INVALID, s.Name, field)
			}
			for operator := range operators {
				if !rollupOperators[operator] {
					return s, fmt.Errorf(MSG_ROLLUP_OPERATOR_INVALID, s.Name, operator)
				}
			}
		}
	}

	s.Wheres = wheres
	return s, nil
}

/**
* NewRollup: Rollup suelto (sin columna en el modelo) para pasarlo a una consulta con Ql.AddRollups o RunQuery.
* Sirve para rollups definidos en datos (no en codigo) sin tocar el modelo compartido mientras se consulta.
* @param name string, with *Model, fks map[string]string, fields []string, show ShowRollup
* @return *Rollup
**/
func NewRollup(name string, with *Model, fks map[string]string, fields []string, show ShowRollup) *Rollup {
	result := &Rollup{
		Name:   name,
		With:   with,
		Fk:     fks,
		Fields: fields,
		Show:   show,
	}
	if with != nil {
		result.WithName = with.Name
	}

	return result
}

/**
* StrToShowRollup: Modo de un rollup desde su texto ("objects"); cualquier otro es attribs.
* @param value string
* @return ShowRollup
**/
func StrToShowRollup(value string) ShowRollup {
	if value == ShowObject.Str() {
		return ShowObject
	}

	return ShowAtrib
}

/**
* getWith: Modelo del que se traen los datos; si no esta cargado (definicion serializada) lo busca por WithName.
* @param db *DB
* @return *Model
**/
func (s *Rollup) getWith(db *DB) *Model {
	if s.With == nil && db != nil && s.WithName != "" {
		s.With = db.GetModel(s.WithName)
	}

	return s.With
}

/**
* Where
* @param from et.Json
* @return et.Json
**/
func (s *Rollup) Where(from et.Json) et.Json {
	result := et.Json{}
	for fkn, pkn := range s.Fk {
		fk := from.Str(fkn)
		result[pkn] = et.Json{
			"eq": fk,
		}
	}

	return result
}

/**
* Serialize
* @return []byte, error
**/
func (s *Rollup) Serialize() ([]byte, error) {
	result, err := json.Marshal(s)
	if err != nil {
		return []byte{}, err
	}

	return result, nil
}

type Join struct {
	On   et.Json  `json:"on"`
	Type TypeJoin `json:"type"`
}

type FullText struct {
	Language string   `json:"language"`
	Columns  []string `json:"columns"`
}

type Column struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	TypeColumn  TypeColumn  `json:"type_column"`
	TypeData    TypeData    `json:"type_data"`
	Default     interface{} `json:"default"`
	Max         float64     `json:"max"`
	Min         float64     `json:"min"`
	Hidden      bool        `json:"hidden"`
	model       *Model      `json:"-"`
}

func init() {
	gob.Register(&Column{})
}

/**
* newColumn
* @param model *Model, name string, description string, typeColumn TypeColumn, typeData TypeData, def interface{}
* @return *Column
**/
func newColumn(model *Model, name string, description string, typeColumn TypeColumn, typeData TypeData, def interface{}) *Column {
	return &Column{
		model:       model,
		Name:        name,
		Description: description,
		TypeColumn:  typeColumn,
		TypeData:    typeData,
		Default:     def,
		Max:         0,
		Min:         0,
	}
}

/**
* newAtribute
* @param model *Model, name string, typeData TypeData
* @return *Column
**/
func newAtribute(model *Model, name string, typeData TypeData) *Column {
	if model.SourceField == nil {
		return nil
	}

	result := model.getColumn(name)
	if result != nil {
		return result
	}

	result = newColumn(model, name, "", TpAtribute, typeData, typeData.DefaultValue())
	return result
}

/**
* Describe
* @return et.Json
**/
func (s *Column) Describe() et.Json {
	result := et.Json{
		"name":        s.Name,
		"description": s.Description,
		"type_column": s.TypeColumn.Str(),
		"type_data":   s.TypeData.Str(),
		"default":     s.Default,
		"max":         s.Max,
		"min":         s.Min,
		"hidden":      s.Hidden,
	}

	return result
}

/**
* Idx
* @return int
**/
func (s *Column) idx() int {
	if s.model == nil {
		return -1
	}

	return slices.IndexFunc(s.model.Columns, func(e *Column) bool { return e.Name == s.Name })
}

/**
* SetDefaultValue
* @param value interface{}
* @return *Column
**/
func (s *Column) SetDefaultValue(value interface{}) *Column {
	s.Default = value
	return s
}

/**
* SetHidden
* @param hidden bool
* @return *Column
**/
func (s *Column) SetHidden(hidden bool) *Column {
	s.Hidden = hidden
	return s
}

/**
* SetMax
* @param max float64
* @return *Column
**/
func (s *Column) SetMax(max float64) *Column {
	s.Max = max
	return s
}

/**
* SetMin
* @param min float64
* @return *Column
**/
func (s *Column) SetMin(min float64) *Column {
	s.Min = min
	return s
}

/**
* DefaultValue
* @return interface{}
**/
func (s *Column) DefaultValue() interface{} {
	if s.TypeData == TypeDataDateTime {
		return timezone.Now()
	}

	return s.Default
}

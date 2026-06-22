package inbox

import (
	"fmt"

	"github.com/celsiainternet/elvis/dt"
	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/msg"
	"github.com/celsiainternet/elvis/strs"
	"github.com/celsiainternet/elvis/utility"
	"github.com/celsiainternet/jdb/jdb"
)

type Inbox struct {
	schema *jdb.Schema
	model  *jdb.Model
}

var inb *Inbox

/**
* Load
* @param db *jdb.DB, schema, name string
* @return (*Inbox, error)
**/
func Load(db *jdb.DB, schema string) error {
	if inb != nil {
		return nil
	}

	var err error
	inb, err = Define(db, schema)
	if err != nil {
		return err
	}

	return nil
}

/**
* Define
* @param db *jdb.DB, schema, name string
* @return (*Inbox, error)
**/
func Define(db *jdb.DB, schema string) (*Inbox, error) {
	schemaObj, err := defineSchema(db, schema)
	if err != nil {
		return nil, err
	}

	name := "inboxes"
	model := jdb.NewModel(schemaObj, name, 1)
	model.DefineProjectModel()
	model.DefineColumn("user_id", jdb.TypeDataKey) // Identificador del usuario del sistema
	model.DefineColumn("app_id", jdb.TypeDataKey)
	model.DefineColumn("kind", jdb.TypeDataText)
	model.DefineColumn("code", jdb.TypeDataKey)
	model.DefineColumn("title", jdb.TypeDataText)
	model.DefineColumn("client_id", jdb.TypeDataText) // Identificador del cliente
	model.DefineIndex(true,
		"user_id",
		"app_id",
		"kind",
		"code",
		"title",
		"client_id",
	)
	model.DefineCalc("delete", func(data et.Json) {
		statusId := data.Str(jdb.STATUS_ID)
		if map[string]bool{
			utility.FOR_DELETE: true,
			utility.ARCHIVED:   true,
			utility.CANCELLED:  true,
		}[statusId] {
			data.Set("delete", true)
		} else {
			data.Set("delete", false)
		}
	})
	model.BeforeInsert(func(tx *jdb.Tx, data et.Json) error {
		id := data.Str(jdb.KEY)
		exist, err := model.
			Where(jdb.KEY).Eq(id).
			ItExistsTx(tx)
		if err != nil {
			return err
		}

		if exist {
			return fmt.Errorf(msg.RECORD_NOT_FOUND)
		}

		return nil
	})

	if err := model.Init(); err != nil {
		return nil, err
	}

	return &Inbox{
		schema: schemaObj,
		model:  model,
	}, nil
}

/**
* GetInboxesById
* @param id string
* @return et.Item, error
**/
func (s *Inbox) GetInboxesById(id string) (et.Item, error) {
	key := strs.Format("inbox:%s", id)
	obj := dt.Get(key)
	if obj.Ok {
		return obj.Item, nil
	}

	item, err := s.model.
		Where(jdb.KEY).Eq(id).
		One()
	if err != nil {
		return et.Item{}, err
	}

	dt.Up(key, item)
	return item, nil
}

/**
* GetInboxesByCode
* @param code string
* @return et.Item, error
**/
func (s *Inbox) GetInboxesByCode(kind, code string) (et.Item, error) {
	key := strs.Format("inbox:%s:%s", kind, code)
	obj := dt.Get(key)
	if obj.Ok {
		return obj.Item, nil
	}

	item, err := s.model.
		Where("code").Eq(code).
		One()
	if err != nil {
		return et.Item{}, err
	}

	dt.Up(key, item)
	return item, nil
}

/**
* GetInboxesByUserId
* @param userId, appId, kind, status string, page, limit int
* @return et.Items, error
**/
func (s *Inbox) GetInboxesByUserId(userId, appId, kind, status string, page, rows int) (et.Items, error) {
	ql := s.model.
		Where("user_id").Eq(userId).
		And("app_id").Eq(appId).
		And("kind").Eq(kind)
	if status == "0" {
		ql = ql.And("_status").In(status, "-2")
	} else {
		ql = ql.And("_status").Eq(status)
	}

	result, err := ql.
		OrderByDesc(jdb.UPDATED_AT).
		Page(page).
		Rows(rows)
	if err != nil {
		return et.Items{}, err
	}

	return result, nil
}

/**
* GetInboxesByClientId
* @param clientId, appId, status string, page, rows int
* @return et.Items, error
**/
func (s *Inbox) GetInboxesByClientId(clientId, appId, status string, page, rows int) (et.Items, error) {
	ql := s.model.
		Where("client_id").Eq(clientId).
		And("app_id").Eq(appId)
	if status == "0" {
		ql = ql.And("_status").In(status, "-2")
	} else {
		ql = ql.And("_status").Eq(status)
	}

	result, err := ql.
		OrderByDesc(jdb.UPDATED_AT).
		Page(page).
		Rows(rows)
	if err != nil {
		return et.Items{}, err
	}

	return result, nil
}

/**
* GenInboxesCode
* @param projectId string
* @return string, error
**/
func (s *Inbox) GenInboxesCode(projectId string) (string, error) {
	code, err := jdb.GetSeries("services", projectId)
	if err != nil {
		return "", err
	}

	return code, nil
}

/**
* UpsertInboxes
* @param projectId, id string, userId, appId, kind string, data et.Json, createdBy string
* @return et.Item, error
**/
func (s *Inbox) UpsertInboxes(projectId, id, clientId, appId, kind string, data et.Json, userId string) (et.Item, error) {
	if !utility.ValidStr(projectId, 0, []string{""}) {
		return et.Item{}, fmt.Errorf(msg.MSG_ATRIB_REQUIRED, jdb.PROJECT_ID)
	}

	if !utility.ValidStr(id, 0, []string{""}) {
		return et.Item{}, fmt.Errorf(msg.MSG_ATRIB_REQUIRED, "id")
	}

	if !utility.ValidStr(clientId, 0, []string{""}) {
		return et.Item{}, fmt.Errorf(msg.MSG_ATRIB_REQUIRED, "client_id")
	}

	if !utility.ValidStr(userId, 0, []string{""}) {
		return et.Item{}, fmt.Errorf(msg.MSG_ATRIB_REQUIRED, "user_id")
	}

	if !utility.ValidStr(appId, 0, []string{""}) {
		return et.Item{}, fmt.Errorf(msg.MSG_ATRIB_REQUIRED, "app_id")
	}

	if !utility.ValidStr(kind, 0, []string{""}) {
		return et.Item{}, fmt.Errorf(msg.MSG_ATRIB_REQUIRED, "kind")
	}

	id = s.model.GetId(id)
	now := utility.Now()
	data[jdb.PROJECT_ID] = projectId
	data[jdb.KEY] = id
	data["kind"] = kind
	result, err := s.model.
		Upsert(data).
		BeforeInsert(func(tx *jdb.Tx, data et.Json) error {
			code := data.Str("code")
			if code == "" {
				code, err := s.GenInboxesCode(projectId)
				if err == nil {
					data["code"] = code
				}
			}
			data[jdb.CREATED_AT] = now
			data[jdb.UPDATED_AT] = now
			data[jdb.STATUS_ID] = utility.ACTIVE
			data["app_id"] = appId
			data["client_id"] = clientId
			data["user_id"] = userId
			auditLog := data.ArrayJson("audit_log")
			auditLog = append(auditLog, et.Json{
				"user_id": userId,
				"action":  "create",
				"date":    now,
			})
			return nil
		}).
		BeforeUpdate(func(tx *jdb.Tx, data et.Json) error {
			data[jdb.UPDATED_AT] = now
			auditLog := data.ArrayJson("audit_log")
			auditLog = append(auditLog, et.Json{
				"user_id": userId,
				"action":  "update",
				"date":    now,
			})
			return nil
		}).
		Where(jdb.STATUS_ID).Eq(utility.ACTIVE).
		Exec()
	if err != nil {
		return et.Item{}, err
	}

	key := strs.Format("inbox:%s", id)
	dt.Drop(key)

	return result.First(), nil
}

/**
* StateInboxes
* @param id, status, userId string
* @return et.Item, error
**/
func (s *Inbox) StateInboxes(id, status, userId string) (et.Item, error) {
	if !utility.ValidStr(status, 0, []string{""}) {
		return et.Item{}, fmt.Errorf(msg.MSG_ATRIB_REQUIRED, jdb.STATUS_ID)
	}

	if !utility.ValidStr(id, 0, []string{""}) {
		return et.Item{}, fmt.Errorf(msg.MSG_ATRIB_REQUIRED, jdb.KEY)
	}

	now := utility.Now()
	result, err := s.model.
		Update(et.Json{
			jdb.STATUS_ID: status,
		}).
		BeforeUpdate(func(tx *jdb.Tx, data et.Json) error {
			data[jdb.UPDATED_AT] = now
			auditLog := data.ArrayJson("audit_log")
			auditLog = append(auditLog, et.Json{
				"user_id": userId,
				"action":  fmt.Sprintf("cambios de estado a %s", status),
				"date":    now,
			})
			return nil
		}).
		Where(jdb.KEY).Eq(id).
		And(jdb.STATUS_ID).Neg(status).
		One()
	if err != nil {
		return et.Item{}, err
	}

	key := strs.Format("inbox:%s", id)
	dt.Drop(key)

	return et.Item{
		Ok: result.Ok,
		Result: et.Json{
			"message": msg.RECORD_UPDATE,
		},
	}, nil
}

/**
* QueryInboxes
* @param query et.Json
* @return interface{}, error
**/
func (s *Inbox) QueryInboxes(query et.Json) (interface{}, error) {
	result, err := jdb.From(s.model).
		Query(query)
	if err != nil {
		return nil, err
	}

	return result, nil
}

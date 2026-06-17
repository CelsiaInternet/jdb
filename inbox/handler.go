package inbox

import (
	"fmt"

	"github.com/celsiainternet/elvis/et"
)

/**
* GetInboxesById
* @param id string
* @return et.Item, error
**/
func GetInboxesById(id string) (et.Item, error) {
	if inb == nil {
		return et.Item{}, fmt.Errorf("inbox not found")
	}

	return inb.GetInboxesById(id)
}

/**
* GetInboxesByCode
* @param kind, code string
* @return et.Item, error
**/
func GetInboxesByCode(kind, code string) (et.Item, error) {
	if inb == nil {
		return et.Item{}, fmt.Errorf("inbox not found")
	}

	return inb.GetInboxesByCode(kind, code)
}

/**
* GetInboxesByMy
* @param userId, appId, kind, status string, page, rows int
* @return et.Items, error
**/
func GetInboxesByUserId(userId, appId, kind, status string, page, rows int) (et.Items, error) {
	if inb == nil {
		return et.Items{}, fmt.Errorf("inbox not found")
	}

	return inb.GetInboxesByUserId(userId, appId, kind, status, page, rows)
}

/**
* GetInboxesByClientId
* @param clientId, appId, status string, page, rows int
* @return et.Items, error
**/
func GetInboxesByClientId(clientId, appId, status string, page, rows int) (et.Items, error) {
	if inb == nil {
		return et.Items{}, fmt.Errorf("inbox not found")
	}

	return inb.GetInboxesByClientId(clientId, appId, status, page, rows)
}

/**
* GenInboxesCode
* @param projectId string
* @return string, error
**/
func GenInboxesCode(projectId string) (string, error) {
	if inb == nil {
		return "", fmt.Errorf("inbox not found")
	}

	return inb.GenInboxesCode(projectId)
}

/**
* UpsertInboxes
* @param projectId, id, clientId, appId, kind string, data et.Json, userId string
* @return et.Item, error
**/
func UpsertInboxes(projectId, id, clientId, appId, kind string, data et.Json, userId string) (et.Item, error) {
	if inb == nil {
		return et.Item{}, fmt.Errorf("inbox not found")
	}

	return inb.UpsertInboxes(projectId, id, clientId, appId, kind, data, userId)
}

/**
* StateInboxes
* @param id, status, userId string
* @return et.Item, error
**/
func StateInboxes(id, status, userId string) (et.Item, error) {
	if inb == nil {
		return et.Item{}, fmt.Errorf("inbox not found")
	}

	return inb.StateInboxes(id, status, userId)
}

/**
* QueryInboxes
* @param query et.Json
* @return interface{}, error
**/
func QueryInboxes(query et.Json) (interface{}, error) {
	if inb == nil {
		return nil, fmt.Errorf("inbox not found")
	}

	return inb.QueryInboxes(query)
}

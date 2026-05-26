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
* GetInboxesByMy
* @param userId, appId, inbox, status string, page, rows int
* @return et.Items, error
**/
func GetInboxesByMy(userId, appId, inbox, status string, page, rows int) (et.Items, error) {
	if inb == nil {
		return et.Items{}, fmt.Errorf("inbox not found")
	}

	return inb.GetInboxesByMy(userId, appId, inbox, status, page, rows)
}

/**
* GetInboxesCode
* @param projectId, inbox string
* @return string, error
**/
func GetInboxesCode(projectId, inbox string) (string, error) {
	if inb == nil {
		return "", fmt.Errorf("inbox not found")
	}

	return inb.GetInboxesCode(projectId, inbox)
}

/**
* UpsertInboxes
* @param projectId, id, userId, appId, inbox string, data et.Json, createdBy string
* @return et.Item, error
**/
func UpsertInboxes(projectId, id, userId, appId, inbox string, data et.Json, createdBy string) (et.Item, error) {
	if inb == nil {
		return et.Item{}, fmt.Errorf("inbox not found")
	}

	return inb.UpsertInboxes(projectId, id, userId, appId, inbox, data, createdBy)
}

/**
* StateInboxes
* @param id, stateId, createdBy string
* @return et.Item, error
**/
func StateInboxes(id, stateId, createdBy string) (et.Item, error) {
	if inb == nil {
		return et.Item{}, fmt.Errorf("inbox not found")
	}

	return inb.StateInboxes(id, stateId, createdBy)
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

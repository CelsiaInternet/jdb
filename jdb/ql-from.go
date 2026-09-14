package jdb

import (
	"fmt"
	"sync"

	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/strs"
	"github.com/celsiainternet/elvis/utility"
)

func helpQl(model *Model) et.Json {
	return et.Json{
		"from": model.Name,
		"data": []interface{}{
			"name",
			"status_id",
			"kinds.name:status",
			et.Json{
				"folders": et.Json{
					"select": []interface{}{
						"name",
					},
					"page": 1,
					"rows": 30,
					"list": true,
				},
			},
		},
		"join": []et.Json{
			{
				"kinds": et.Json{
					"status_id": et.Json{
						"eq": "kinds.id",
					},
				},
				"AND": []et.Json{},
				"OR":  []et.Json{},
			},
		},
		"where": et.Json{
			"status_id": et.Json{
				"eq": "kinds.id",
			},
			"AND": []et.Json{
				{
					"name": et.Json{
						"eq": "v:name",
					},
				},
			},
			"OR": []et.Json{},
		},
		"group_by": []string{"name"},
		"having": et.Json{
			"name": et.Json{
				"eq": "name",
			},
			"AND": []et.Json{},
			"OR":  []et.Json{},
		},
		"order_by": et.Json{
			"ASC":  []string{"name"},
			"DESC": []string{"name"},
		},
		"page":  1,
		"limit": 30,
	}
}

/**
* getForms
* @return []string
**/
func (s *Ql) getForms() []string {
	var result []string
	for _, from := range s.Froms.Froms {
		result = append(result, strs.Format(`%s.%s, %s`, from.Schema, from.Name, from.As))
	}

	return result
}

package jdb

import (
	"strings"

	"github.com/celsiainternet/elvis/et"
)

/**
* validator
* validate this val is a field or basic type
* @return interface{}
**/
func (s *Command) validator(val interface{}) interface{} {
	switch v := val.(type) {
	case string:
		if strings.HasPrefix(v, ":") {
			v = strings.TrimPrefix(v, ":")
			field := s.getField(v)
			if field != nil {
				return field
			}
			return nil
		}

		if strings.HasPrefix(v, "$") {
			v = strings.TrimPrefix(v, "$")
			return v
		}

		v = strings.Replace(v, `\\:`, `\:`, 1)
		v = strings.Replace(v, `\:`, `:`, 1)
		v = strings.Replace(v, `\\$`, `\$`, 1)
		v = strings.Replace(v, `\$`, `$`, 1)
		field := s.getField(v)
		if field != nil {
			return field
		}

		return v
	case *Field:
		return v
	case Field:
		return v
	case *Column:
		return GetField(v)
	case Column:
		return GetField(&v)
	case []interface{}:
		return v
	case []string:
		return v
	case []et.Json:
		return v
	default:
		if v == nil {
			return "nil"
		}

		return v
	}
}

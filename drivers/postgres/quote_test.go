package postgres

import (
	"testing"

	"github.com/celsiainternet/elvis/et"
)

/**
* TestQuoteEscapesJson: Un objeto o una lista con un apóstrofo queda como literal válido ('' dentro del texto), igual que
* un string. Antes se pegaban sin escapar y el apóstrofo cerraba el literal (error de sintaxis o inyección SQL).
* @param t *testing.T
**/
func TestQuoteEscapesJson(t *testing.T) {
	cases := []struct {
		name  string
		value any
		want  string
	}{
		{"string", "Et library's", `'Et library''s'`},
		{"et.Json", et.Json{"code": "library's"}, `'{"code":"library''s"}'`},
		{"map", map[string]interface{}{"code": "library's"}, `'{"code":"library''s"}'`},
		{"lista", []string{"it's"}, `'["it''s"]'`},
	}

	for _, c := range cases {
		if got := quote(c.value); got != c.want {
			t.Errorf("%s: got %v, want %s", c.name, got, c.want)
		}
	}
}

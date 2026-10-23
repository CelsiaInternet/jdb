package jdb

import (
	"regexp"
	"strings"
)

var (
	// Postgres: columna seguida de ->'llave' o ->>'llave' (ej. _data->'sync_external'->>'oss_order_id').
	jsonArrowPath    = regexp.MustCompile(`^\s*([A-Za-z_][A-Za-z0-9_]*)((?:\s*->>?\s*'[^']*')+)\s*$`)
	jsonArrowSegment = regexp.MustCompile(`->>?\s*'([^']*)'`)
	// Cada llave de la ruta: letras, numeros, _ y -. Evita inyectar SQL en la ruta.
	jsonPathKey = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
)

/**
* splitJsonPath: Separa un nombre de campo con ruta JSON en su primer segmento y las llaves de la ruta.
* Acepta la sintaxis de Postgres (_data->'a'->>'b') y la de puntos (_data.a.b). Devuelve ok = false
* si el nombre no es una ruta o alguna llave tiene caracteres no permitidos.
* @param name string
* @return string, []string, bool
**/
func splitJsonPath(name string) (string, []string, bool) {
	var segments []string
	if match := jsonArrowPath.FindStringSubmatch(name); match != nil {
		segments = append(segments, match[1])
		for _, key := range jsonArrowSegment.FindAllStringSubmatch(match[2], -1) {
			segments = append(segments, key[1])
		}
	} else if strings.Contains(name, ".") {
		segments = strings.Split(strings.TrimSpace(name), ".")
	}

	if len(segments) < 2 {
		return "", nil, false
	}

	for _, key := range segments {
		if !jsonPathKey.MatchString(key) {
			return "", nil, false
		}
	}

	return segments[0], segments[1:], true
}

/**
* getPathField: Resuelve un nombre con ruta JSON a un campo con Path.
* Si el primer segmento es una columna JSON (TypeDataObject) del modelo, la ruta va dentro de ella
* (_data.sync_external.oss_order_id); si no y el modelo tiene source, la ruta completa va dentro del source
* (sync_external.oss_order_id). Devuelve nil si el nombre no es una ruta o no aplica a este modelo.
* @param name string
* @return *Field
**/
func (s *Model) getPathField(name string) *Field {
	first, path, ok := splitJsonPath(name)
	if !ok {
		return nil
	}

	column := s.getColumn(first)
	if column != nil {
		if column.TypeColumn != TpColumn || column.TypeData != TypeDataObject {
			return nil
		}

		result := GetField(column)
		result.Path = path
		result.As = path[len(path)-1]
		return result
	}

	if s.SourceField == nil {
		return nil
	}

	result := GetField(s.SourceField)
	result.Path = append([]string{first}, path...)
	result.As = path[len(path)-1]
	return result
}

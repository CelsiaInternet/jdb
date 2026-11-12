package jdb

import (
	"slices"
	"strings"
)

/* SEARCH_MAX_WORDS: palabras que toma una busqueda; las demas se ignoran para no armar un SQL enorme. */
const SEARCH_MAX_WORDS = 10

/**
* TextSearch: Los campos de busqueda de un modelo (DefineSearch) unidos con un espacio, como un solo
* texto. Cada palabra buscada es una condicion Search sobre este campo:
*
*	ql.WhereSearch("ana gomez")   // ... texto LIKE '%ana%' AND texto LIKE '%gomez%'
*
* En postgres el texto va sin tildes y en minusculas, con la misma expresion del indice de trigramas
* que crea el driver, para que la consulta lo use. En SQLite y Oracle es un LIKE sin indice.
**/
type TextSearch struct {
	Fields []*Field `json:"fields"`
}

/**
* DefineSearch: Declara los campos (columnas o atributos ya definidos, de texto) que entran en la busqueda
* del modelo (WhereSearch, "search" en RunQuery). Se ignoran los nombres que no son campos y las fechas. En postgres el
* driver crea, al iniciar el modelo, el indice de trigramas sobre esos campos.
* @param fields ...string
* @return *Model
**/
func (s *Model) DefineSearch(fields ...string) *Model {
	for _, name := range fields {
		col := s.getColumn(name)
		if col == nil {
			continue
		}

		if col.TypeColumn != TpColumn && col.TypeColumn != TpAtribute {
			continue
		}

		// Una fecha como texto depende de la configuracion de la sesion: postgres no la deja en un indice.
		if col.TypeData == TypeDataDateTime {
			continue
		}

		if slices.Contains(s.SearchFields, name) {
			continue
		}

		s.SearchFields = append(s.SearchFields, name)
	}

	return s
}

/**
* SearchWords: Palabras de text sin repetir (sin importar mayusculas), hasta SEARCH_MAX_WORDS.
* @param text string
* @return []string
**/
func SearchWords(text string) []string {
	result := []string{}
	seen := map[string]bool{}
	for _, word := range strings.Fields(text) {
		key := strings.ToLower(word)
		if seen[key] {
			continue
		}

		seen[key] = true
		result = append(result, word)
		if len(result) == SEARCH_MAX_WORDS {
			break
		}
	}

	return result
}

/**
* WhereSearch: Agrega una condicion por palabra de text sobre los campos de busqueda del modelo
* (DefineSearch); todas deben aparecer. Sin palabras no agrega nada. Si el modelo no declaro campos,
* la condicion igual se agrega y no encuentra nada (como WhereConcat sin campos).
* @param text string
* @return *Ql
**/
func (s *Ql) WhereSearch(text string) *Ql {
	words := SearchWords(text)
	if len(words) == 0 {
		return s
	}

	fields := []*Field{}
	if s.Froms != nil && len(s.Froms.Froms) > 0 {
		from := s.Froms.Froms[0]
		for _, name := range from.SearchFields {
			field := s.getField(name)
			if field != nil {
				fields = append(fields, field)
			}
		}
	}

	for _, word := range words {
		search := &TextSearch{Fields: fields}
		s.Where(search).Search(word)
	}

	return s
}

/**
* WhereSearch: Inicia una consulta del modelo con la busqueda de text en sus campos de busqueda.
* @param text string
* @return *Ql
**/
func (s *Model) WhereSearch(text string) *Ql {
	result := From(s)
	if s.SourceField != nil {
		result.TypeSelect = Source
	}

	return result.WhereSearch(text)
}

/**
* Search: Operador Search sobre la condicion abierta: el campo contiene val (en postgres sin distinguir
* mayusculas; sobre un TextSearch, tampoco tildes).
* @param val interface{}
* @return *Ql
**/
func (s *Ql) Search(val interface{}) *Ql {
	s.QlWhere.Search(val)
	return s
}

/**
* Search: Operador Search sobre la ultima condicion.
* @param val interface{}
* @return *QlWhere
**/
func (s *QlWhere) Search(val interface{}) *QlWhere {
	condition := s.condition()
	if condition == nil {
		return s
	}

	condition.Operator = Search
	condition.setValue(val)
	return s
}

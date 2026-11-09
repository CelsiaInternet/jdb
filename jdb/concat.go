package jdb

/**
* Concat: Varios campos unidos con un espacio, como un solo texto. Sirve para buscar un valor en
* todos a la vez con una sola condicion, por ejemplo
*
*	ql.WhereConcat("full_name", "email").Like("%ana%")
*
* A diferencia de un OR por campo, es una condicion: no rompe el AND con las demas (alcance,
* filtros) aunque los drivers no pongan parentesis. Un campo nulo cuenta como texto vacio.
**/
type Concat struct {
	Fields []*Field `json:"fields"`
}

/**
* NewConcat: Concat con los campos dados; ignora los nil.
* @param fields ...*Field
* @return *Concat
**/
func NewConcat(fields ...*Field) *Concat {
	result := &Concat{Fields: []*Field{}}
	for _, field := range fields {
		if field != nil {
			result.Fields = append(result.Fields, field)
		}
	}

	return result
}

/**
* WhereConcat: Agrega una condicion sobre los campos dados unidos con un espacio (Concat). Los
* nombres que no son campos del modelo se ignoran; si no queda ninguno la condicion igual se agrega
* (como texto vacio), para que el operador que siga no caiga sobre la condicion anterior.
* @param names ...string
* @return *Ql
**/
func (s *Ql) WhereConcat(names ...string) *Ql {
	fields := []*Field{}
	for _, name := range names {
		if field := s.getField(name); field != nil {
			fields = append(fields, field)
		}
	}

	return s.Where(NewConcat(fields...))
}

/**
* WhereConcat: Inicia una consulta del modelo con una condicion sobre varios campos unidos.
* @param names ...string
* @return *Ql
**/
func (s *Model) WhereConcat(names ...string) *Ql {
	result := From(s)
	if s.SourceField != nil {
		result.TypeSelect = Source
	}

	return result.WhereConcat(names...)
}

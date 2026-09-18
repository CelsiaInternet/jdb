package jdb

/**
* GroupBy
* @param fields ...string
* @return *Ql
**/
func (s *Ql) GroupBy(fields ...string) *Ql {
	if s.Groups == nil {
		s.Groups = []*Field{}
	}
	
	for _, field := range fields {
		field := s.getField(field)
		if field != nil {
			s.Groups = append(s.Groups, field)
		}
	}

	return s
}

/**
* setGroupBy
* @param fields ...string
* @return *Ql
**/
func (s *Ql) setGroupBy(fields ...string) *Ql {
	if len(fields) == 0 {
		return s
	}

	return s.GroupBy(fields...)
}

/**
* getGroupsBy
* @return []string
**/
func (s *Ql) getGroupsBy() []string {
	result := []string{}
	for _, field := range s.Groups {
		def := field.asName()
		result = append(result, def)
	}

	return result
}

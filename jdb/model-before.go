package jdb

/**
* BeforeInsert
* @param fn DataFunction
**/
func (s *Model) BeforeInsert(fn DataFunctionTx) *Model {
	s.beforeInsert = append(s.beforeInsert, fn)

	return s
}

/**
* BeforeUpdate
* @param fn DataFunction
**/
func (s *Model) BeforeUpdate(fn DataFunctionTx) *Model {
	s.beforeUpdate = append(s.beforeUpdate, fn)

	return s
}

/**
* BeforeDelete
* @param fn DataFunction
**/
func (s *Model) BeforeDelete(fn DataFunctionTx) *Model {
	s.beforeDelete = append(s.beforeDelete, fn)

	return s
}

/**
* BeforeInsertOrUpdate
* @param fn DataFunction
**/
func (s *Model) BeforeInsertOrUpdate(fn DataFunctionTx) *Model {
	s.beforeInsert = append(s.beforeInsert, fn)
	s.beforeUpdate = append(s.beforeUpdate, fn)

	return s
}

/**
* BeforeInsertTrigger
* @param fn TriggerFunction
**/
func (s *Model) BeforeInsertTrigger(fn TriggerFunctionTx) *Model {
	s.beforeInsertTrigger = append(s.beforeInsertTrigger, fn)

	return s
}

/**
* BeforeUpdateTrigger
* @param fn TriggerFunction
**/
func (s *Model) BeforeUpdateTrigger(fn TriggerFunctionTx) *Model {
	s.beforeUpdateTrigger = append(s.beforeUpdateTrigger, fn)

	return s
}

/**
* BeforeDeleteTrigger
* @param fn TriggerFunction
**/
func (s *Model) BeforeDeleteTrigger(fn TriggerFunctionTx) *Model {
	s.beforeDeleteTrigger = append(s.beforeDeleteTrigger, fn)

	return s
}

/**
* AfterInsertTrigger
* @param fn TriggerFunction
**/
func (s *Model) BeforeInsertOrUpdateTrigger(fn TriggerFunctionTx) *Model {
	s.beforeInsertTrigger = append(s.beforeInsertTrigger, fn)
	s.beforeUpdateTrigger = append(s.beforeUpdateTrigger, fn)

	return s
}

package jdb

/**
* AfterInsert
* @param fn DataFunction
* @return *Command
**/
func (s *Command) AfterInsert(fn DataFunctionTx) *Command {
	s.afterInsert = append(s.afterInsert, fn)

	return s
}

/**
* AfterUpdate
* @param fn DataFunctionTx
* @return *Command
**/
func (s *Command) AfterUpdate(fn DataFunctionTx) *Command {
	s.afterUpdate = append(s.afterUpdate, fn)

	return s
}

/**
* AfterDelete
* @param fn DataFunctionTx
* @return *Command
**/
func (s *Command) AfterDelete(fn DataFunctionTx) *Command {
	s.afterDelete = append(s.afterDelete, fn)

	return s
}

/**
* AfterInsertOrUpdate
* @param fn DataFunctionTx
* @return *Command
**/
func (s *Command) AfterInsertOrUpdate(fn DataFunctionTx) *Command {
	s.afterInsert = append(s.afterInsert, fn)
	s.afterUpdate = append(s.afterUpdate, fn)

	return s
}

/**
* AfterInsertTrigger
* @param fn TriggerFunction
**/
func (s *Command) AfterInsertTrigger(fn TriggerFunctionTx) *Command {
	s.afterInsertTrigger = append(s.afterInsertTrigger, fn)

	return s
}

/**
* AfterUpdateTrigger
* @param fn TriggerFunction
**/
func (s *Command) AfterUpdateTrigger(fn TriggerFunctionTx) *Command {
	s.afterUpdateTrigger = append(s.afterUpdateTrigger, fn)

	return s
}

/**
* AfterDeleteTrigger
* @param fn TriggerFunction
**/
func (s *Command) AfterDeleteTrigger(fn TriggerFunctionTx) *Command {
	s.afterDeleteTrigger = append(s.afterDeleteTrigger, fn)

	return s
}

/**
* AfterInsertOrUpdateTrigger
* @param fn TriggerFunction
**/
func (s *Command) AfterInsertOrUpdateTrigger(fn TriggerFunctionTx) *Command {
	s.afterInsertTrigger = append(s.afterInsertTrigger, fn)
	s.afterUpdateTrigger = append(s.afterUpdateTrigger, fn)

	return s
}

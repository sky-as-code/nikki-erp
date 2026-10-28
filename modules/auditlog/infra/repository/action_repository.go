package repository

import (
	itAction "github.com/sky-as-code/nikki-erp/modules/auditlog/interfaces/action"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewActionRepository is handed the composable default by the auditlog_action onion. The
// default insert joins the transaction the context carries, which is what puts an entry in the
// same transaction as the change it records.
func NewActionRepository(base composable.CrudRepository) itAction.ActionRepository {
	return &ActionRepositoryImpl{CrudRepository: base}
}

type ActionRepositoryImpl struct {
	composable.CrudRepository
}

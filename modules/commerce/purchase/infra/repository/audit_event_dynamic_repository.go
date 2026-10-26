package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itAudit "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/audit"
)

func NewAuditEventRepository(base composable.CrudRepository) itAudit.AuditEventCrudRepository {
	return &AuditEventRepositoryImpl{CrudRepository: base}
}

type AuditEventRepositoryImpl struct {
	composable.CrudRepository
}

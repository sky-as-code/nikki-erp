package repository

import (
	itAudit "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/audit"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

func NewAuditEventRepository(base composable.CrudRepository) itAudit.AuditEventCrudRepository {
	return &AuditEventRepositoryImpl{CrudRepository: base}
}

type AuditEventRepositoryImpl struct {
	composable.CrudRepository
}

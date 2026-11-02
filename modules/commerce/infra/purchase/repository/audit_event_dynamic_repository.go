package repository

import (
	itAudit "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/audit"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

func NewAuditEventRepository(base dynamicresource.CrudRepository) itAudit.AuditEventCrudRepository {
	return &AuditEventRepositoryImpl{CrudRepository: base}
}

type AuditEventRepositoryImpl struct {
	dynamicresource.CrudRepository
}

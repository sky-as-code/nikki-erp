package services

import (
	itAction "github.com/sky-as-code/nikki-erp/modules/auditlog/interfaces/action"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewActionDomainService is handed the composable default by the auditlog_action onion, which is
// built with IsAudited false: recording an entry must not record another.
func NewActionDomainService(base composable.CrudDomainService) itAction.ActionDomainService {
	return &ActionDomainServiceImpl{CrudDomainService: base}
}

type ActionDomainServiceImpl struct {
	composable.CrudDomainService
}

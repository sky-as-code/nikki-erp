package services

import (
	itWarehouse "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/warehouse"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewPutawayRuleDomainService is handed the composable default by the putaway_rule onion. The resource is plain
// CRUD; the type exists so a rule has a home when one is needed.
func NewPutawayRuleDomainService(base composable.CrudDomainService) itWarehouse.PutawayRuleDomainService {
	return &PutawayRuleDomainServiceImpl{CrudDomainService: base}
}

type PutawayRuleDomainServiceImpl struct {
	composable.CrudDomainService
}

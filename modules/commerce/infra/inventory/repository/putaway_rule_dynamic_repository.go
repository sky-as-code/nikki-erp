package repository

import (
	itWarehouse "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/warehouse"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// NewPutawayRuleRepository is handed the composable default by the putaway_rule onion. The type exists so a
// bespoke query has a home; it adds nothing today.
func NewPutawayRuleRepository(base dynamicresource.CrudRepository) itWarehouse.PutawayRuleRepository {
	return &PutawayRuleRepositoryImpl{CrudRepository: base}
}

type PutawayRuleRepositoryImpl struct {
	dynamicresource.CrudRepository
}

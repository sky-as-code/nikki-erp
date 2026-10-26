package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itWarehouse "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/warehouse"
)

// NewPutawayRuleRepository is handed the composable default by the putaway_rule onion. The type exists so a
// bespoke query has a home; it adds nothing today.
func NewPutawayRuleRepository(base composable.CrudRepository) itWarehouse.PutawayRuleRepository {
	return &PutawayRuleRepositoryImpl{CrudRepository: base}
}

type PutawayRuleRepositoryImpl struct {
	composable.CrudRepository
}

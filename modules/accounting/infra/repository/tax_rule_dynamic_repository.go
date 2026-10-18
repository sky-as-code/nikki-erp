package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewTaxRuleRepository is handed the composable default by the onion.
func NewTaxRuleRepository(base composable.CrudRepository) taxconfig.TaxRuleRepository {
	return &TaxRuleRepositoryImpl{CrudRepository: base}
}

type TaxRuleRepositoryImpl struct {
	composable.CrudRepository
}

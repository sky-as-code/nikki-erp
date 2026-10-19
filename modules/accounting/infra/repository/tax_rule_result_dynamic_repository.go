package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewTaxRuleResultRepository is handed the composable default by the onion.
func NewTaxRuleResultRepository(base composable.CrudRepository) taxconfig.TaxRuleResultRepository {
	return &TaxRuleResultRepositoryImpl{CrudRepository: base}
}

type TaxRuleResultRepositoryImpl struct {
	composable.CrudRepository
}

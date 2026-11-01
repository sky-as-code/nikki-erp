package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// NewTaxRuleResultRepository is handed the composable default by the onion.
func NewTaxRuleResultRepository(base dynamicresource.CrudRepository) taxconfig.TaxRuleResultRepository {
	return &TaxRuleResultRepositoryImpl{CrudRepository: base}
}

type TaxRuleResultRepositoryImpl struct {
	dynamicresource.CrudRepository
}

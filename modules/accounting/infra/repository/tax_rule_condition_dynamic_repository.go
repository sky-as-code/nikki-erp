package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// NewTaxRuleConditionRepository is handed the composable default by the onion.
func NewTaxRuleConditionRepository(base dynamicresource.CrudRepository) taxconfig.TaxRuleConditionRepository {
	return &TaxRuleConditionRepositoryImpl{CrudRepository: base}
}

type TaxRuleConditionRepositoryImpl struct {
	dynamicresource.CrudRepository
}

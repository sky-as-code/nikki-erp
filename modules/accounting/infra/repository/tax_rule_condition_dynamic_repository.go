package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewTaxRuleConditionRepository is handed the composable default by the onion.
func NewTaxRuleConditionRepository(base composable.CrudRepository) taxconfig.TaxRuleConditionRepository {
	return &TaxRuleConditionRepositoryImpl{CrudRepository: base}
}

type TaxRuleConditionRepositoryImpl struct {
	composable.CrudRepository
}

package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewTaxRoundingPolicyRepository is handed the composable default by the onion.
func NewTaxRoundingPolicyRepository(base composable.CrudRepository) taxconfig.TaxRoundingPolicyRepository {
	return &TaxRoundingPolicyRepositoryImpl{CrudRepository: base}
}

type TaxRoundingPolicyRepositoryImpl struct {
	composable.CrudRepository
}

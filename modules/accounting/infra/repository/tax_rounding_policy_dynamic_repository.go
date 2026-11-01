package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// NewTaxRoundingPolicyRepository is handed the composable default by the onion.
func NewTaxRoundingPolicyRepository(base dynamicresource.CrudRepository) taxconfig.TaxRoundingPolicyRepository {
	return &TaxRoundingPolicyRepositoryImpl{CrudRepository: base}
}

type TaxRoundingPolicyRepositoryImpl struct {
	dynamicresource.CrudRepository
}

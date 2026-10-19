package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewTaxRateVersionRepository is handed the composable default by the onion.
func NewTaxRateVersionRepository(base composable.CrudRepository) taxconfig.TaxRateVersionRepository {
	return &TaxRateVersionRepositoryImpl{CrudRepository: base}
}

type TaxRateVersionRepositoryImpl struct {
	composable.CrudRepository
}

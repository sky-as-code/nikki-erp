package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// NewTaxRateVersionRepository is handed the composable default by the onion.
func NewTaxRateVersionRepository(base dynamicresource.CrudRepository) taxconfig.TaxRateVersionRepository {
	return &TaxRateVersionRepositoryImpl{CrudRepository: base}
}

type TaxRateVersionRepositoryImpl struct {
	dynamicresource.CrudRepository
}

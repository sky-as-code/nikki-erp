package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// NewTaxComponentRepository is handed the composable default by the onion.
func NewTaxComponentRepository(base dynamicresource.CrudRepository) taxconfig.TaxComponentRepository {
	return &TaxComponentRepositoryImpl{CrudRepository: base}
}

type TaxComponentRepositoryImpl struct {
	dynamicresource.CrudRepository
}

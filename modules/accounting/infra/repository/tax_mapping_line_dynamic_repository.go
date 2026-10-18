package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewTaxMappingLineRepository is handed the composable default by the onion.
func NewTaxMappingLineRepository(base composable.CrudRepository) taxconfig.TaxMappingLineRepository {
	return &TaxMappingLineRepositoryImpl{CrudRepository: base}
}

type TaxMappingLineRepositoryImpl struct {
	composable.CrudRepository
}

package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewTaxMappingRepository is handed the composable default by the onion.
func NewTaxMappingRepository(base composable.CrudRepository) taxconfig.TaxMappingRepository {
	return &TaxMappingRepositoryImpl{CrudRepository: base}
}

type TaxMappingRepositoryImpl struct {
	composable.CrudRepository
}

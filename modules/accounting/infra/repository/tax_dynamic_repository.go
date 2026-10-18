package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewTaxRepository is handed the composable default by the onion.
func NewTaxRepository(base composable.CrudRepository) taxconfig.TaxRepository {
	return &TaxRepositoryImpl{CrudRepository: base}
}

type TaxRepositoryImpl struct {
	composable.CrudRepository
}

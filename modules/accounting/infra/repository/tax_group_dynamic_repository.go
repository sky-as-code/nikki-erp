package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewTaxGroupRepository is handed the composable default by the onion.
func NewTaxGroupRepository(base composable.CrudRepository) taxconfig.TaxGroupRepository {
	return &TaxGroupRepositoryImpl{CrudRepository: base}
}

type TaxGroupRepositoryImpl struct {
	composable.CrudRepository
}

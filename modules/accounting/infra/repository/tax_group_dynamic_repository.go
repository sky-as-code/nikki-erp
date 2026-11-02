package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// NewTaxGroupRepository is handed the composable default by the onion.
func NewTaxGroupRepository(base dynamicresource.CrudRepository) taxconfig.TaxGroupRepository {
	return &TaxGroupRepositoryImpl{CrudRepository: base}
}

type TaxGroupRepositoryImpl struct {
	dynamicresource.CrudRepository
}

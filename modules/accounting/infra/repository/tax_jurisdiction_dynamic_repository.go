package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewTaxJurisdictionRepository is handed the composable default by the onion.
func NewTaxJurisdictionRepository(base composable.CrudRepository) taxconfig.TaxJurisdictionRepository {
	return &TaxJurisdictionRepositoryImpl{CrudRepository: base}
}

type TaxJurisdictionRepositoryImpl struct {
	composable.CrudRepository
}

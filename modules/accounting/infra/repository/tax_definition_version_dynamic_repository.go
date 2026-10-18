package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewTaxDefinitionVersionRepository is handed the composable default by the onion.
func NewTaxDefinitionVersionRepository(base composable.CrudRepository) taxconfig.TaxDefinitionVersionRepository {
	return &TaxDefinitionVersionRepositoryImpl{CrudRepository: base}
}

type TaxDefinitionVersionRepositoryImpl struct {
	composable.CrudRepository
}

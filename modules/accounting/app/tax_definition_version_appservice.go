package app

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// NewTaxDefinitionVersionApplicationService is handed the composable default by the onion: the authorized CRUD.
func NewTaxDefinitionVersionApplicationService(
	base dynamicresource.CrudApplicationService,
) taxconfig.TaxDefinitionVersionApplicationService {
	return &TaxDefinitionVersionApplicationServiceImpl{CrudApplicationService: base}
}

type TaxDefinitionVersionApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
}

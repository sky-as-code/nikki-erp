package app

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// NewTaxMappingApplicationService is handed the composable default by the onion: the authorized CRUD.
func NewTaxMappingApplicationService(
	base dynamicresource.CrudApplicationService,
) taxconfig.TaxMappingApplicationService {
	return &TaxMappingApplicationServiceImpl{CrudApplicationService: base}
}

type TaxMappingApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
}

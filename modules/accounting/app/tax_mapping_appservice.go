package app

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewTaxMappingApplicationService is handed the composable default by the onion: the authorized CRUD.
func NewTaxMappingApplicationService(
	base composable.CrudApplicationService,
) taxconfig.TaxMappingApplicationService {
	return &TaxMappingApplicationServiceImpl{CrudApplicationService: base}
}

type TaxMappingApplicationServiceImpl struct {
	composable.CrudApplicationService
}

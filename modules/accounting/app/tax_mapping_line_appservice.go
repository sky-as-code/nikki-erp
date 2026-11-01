package app

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// NewTaxMappingLineApplicationService is handed the composable default by the onion: the authorized CRUD.
func NewTaxMappingLineApplicationService(
	base dynamicresource.CrudApplicationService,
) taxconfig.TaxMappingLineApplicationService {
	return &TaxMappingLineApplicationServiceImpl{CrudApplicationService: base}
}

type TaxMappingLineApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
}

package app

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewTaxMappingLineApplicationService is handed the composable default by the onion: the authorized CRUD.
func NewTaxMappingLineApplicationService(
	base composable.CrudApplicationService,
) taxconfig.TaxMappingLineApplicationService {
	return &TaxMappingLineApplicationServiceImpl{CrudApplicationService: base}
}

type TaxMappingLineApplicationServiceImpl struct {
	composable.CrudApplicationService
}

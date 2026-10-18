package app

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewTaxComponentApplicationService is handed the composable default by the onion: the authorized CRUD.
func NewTaxComponentApplicationService(
	base composable.CrudApplicationService,
) taxconfig.TaxComponentApplicationService {
	return &TaxComponentApplicationServiceImpl{CrudApplicationService: base}
}

type TaxComponentApplicationServiceImpl struct {
	composable.CrudApplicationService
}

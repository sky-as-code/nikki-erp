package app

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewTaxRateVersionApplicationService is handed the composable default by the onion: the authorized CRUD.
func NewTaxRateVersionApplicationService(
	base composable.CrudApplicationService,
) taxconfig.TaxRateVersionApplicationService {
	return &TaxRateVersionApplicationServiceImpl{CrudApplicationService: base}
}

type TaxRateVersionApplicationServiceImpl struct {
	composable.CrudApplicationService
}

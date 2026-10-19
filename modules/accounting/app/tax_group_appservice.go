package app

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewTaxGroupApplicationService is handed the composable default by the onion: the authorized CRUD.
func NewTaxGroupApplicationService(
	base composable.CrudApplicationService,
) taxconfig.TaxGroupApplicationService {
	return &TaxGroupApplicationServiceImpl{CrudApplicationService: base}
}

type TaxGroupApplicationServiceImpl struct {
	composable.CrudApplicationService
}

package app

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewTaxProductClassificationApplicationService is handed the composable default by the onion: the authorized CRUD.
func NewTaxProductClassificationApplicationService(
	base composable.CrudApplicationService,
) taxconfig.TaxProductClassificationApplicationService {
	return &TaxProductClassificationApplicationServiceImpl{CrudApplicationService: base}
}

type TaxProductClassificationApplicationServiceImpl struct {
	composable.CrudApplicationService
}

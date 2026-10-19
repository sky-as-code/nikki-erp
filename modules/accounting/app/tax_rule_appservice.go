package app

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewTaxRuleApplicationService is handed the composable default by the onion: the authorized CRUD.
func NewTaxRuleApplicationService(
	base composable.CrudApplicationService,
) taxconfig.TaxRuleApplicationService {
	return &TaxRuleApplicationServiceImpl{CrudApplicationService: base}
}

type TaxRuleApplicationServiceImpl struct {
	composable.CrudApplicationService
}

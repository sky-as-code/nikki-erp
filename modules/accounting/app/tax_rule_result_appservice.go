package app

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewTaxRuleResultApplicationService is handed the composable default by the onion: the authorized CRUD.
func NewTaxRuleResultApplicationService(
	base composable.CrudApplicationService,
) taxconfig.TaxRuleResultApplicationService {
	return &TaxRuleResultApplicationServiceImpl{CrudApplicationService: base}
}

type TaxRuleResultApplicationServiceImpl struct {
	composable.CrudApplicationService
}

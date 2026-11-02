package app

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// NewTaxRuleApplicationService is handed the composable default by the onion: the authorized CRUD.
func NewTaxRuleApplicationService(
	base dynamicresource.CrudApplicationService,
) taxconfig.TaxRuleApplicationService {
	return &TaxRuleApplicationServiceImpl{CrudApplicationService: base}
}

type TaxRuleApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
}

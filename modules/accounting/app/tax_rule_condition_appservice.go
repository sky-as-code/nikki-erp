package app

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// NewTaxRuleConditionApplicationService is handed the composable default by the onion: the authorized CRUD.
func NewTaxRuleConditionApplicationService(
	base dynamicresource.CrudApplicationService,
) taxconfig.TaxRuleConditionApplicationService {
	return &TaxRuleConditionApplicationServiceImpl{CrudApplicationService: base}
}

type TaxRuleConditionApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
}

package app

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewTaxRoundingPolicyApplicationService is handed the composable default by the onion: the authorized CRUD.
func NewTaxRoundingPolicyApplicationService(
	base composable.CrudApplicationService,
) taxconfig.TaxRoundingPolicyApplicationService {
	return &TaxRoundingPolicyApplicationServiceImpl{CrudApplicationService: base}
}

type TaxRoundingPolicyApplicationServiceImpl struct {
	composable.CrudApplicationService
}

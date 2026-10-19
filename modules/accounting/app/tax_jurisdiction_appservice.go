package app

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewTaxJurisdictionApplicationService is handed the composable default by the onion: the authorized CRUD.
func NewTaxJurisdictionApplicationService(
	base composable.CrudApplicationService,
) taxconfig.TaxJurisdictionApplicationService {
	return &TaxJurisdictionApplicationServiceImpl{CrudApplicationService: base}
}

type TaxJurisdictionApplicationServiceImpl struct {
	composable.CrudApplicationService
}

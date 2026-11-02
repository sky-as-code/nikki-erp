package app

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// NewTaxJurisdictionApplicationService is handed the composable default by the onion: the authorized CRUD.
func NewTaxJurisdictionApplicationService(
	base dynamicresource.CrudApplicationService,
) taxconfig.TaxJurisdictionApplicationService {
	return &TaxJurisdictionApplicationServiceImpl{CrudApplicationService: base}
}

type TaxJurisdictionApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
}

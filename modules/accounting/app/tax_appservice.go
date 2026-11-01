package app

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// NewTaxApplicationService is handed the composable default by the onion: the authorized CRUD.
func NewTaxApplicationService(base dynamicresource.CrudApplicationService) taxconfig.TaxApplicationService {
	return &TaxApplicationServiceImpl{CrudApplicationService: base}
}

type TaxApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
}

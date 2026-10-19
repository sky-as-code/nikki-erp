package app

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewTaxApplicationService is handed the composable default by the onion: the authorized CRUD.
func NewTaxApplicationService(base composable.CrudApplicationService) taxconfig.TaxApplicationService {
	return &TaxApplicationServiceImpl{CrudApplicationService: base}
}

type TaxApplicationServiceImpl struct {
	composable.CrudApplicationService
}

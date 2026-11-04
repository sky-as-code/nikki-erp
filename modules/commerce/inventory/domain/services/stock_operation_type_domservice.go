package services

import (
	itExt "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/external"
	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// NewStockOperationTypeDomainService is handed the composable default by the stock_operation_type
// onion, the Sequence port its transfers are numbered through and the reader of the org setting
// naming the definition.
func NewStockOperationTypeDomainService(
	base dynamicresource.CrudDomainService, sequences itExt.SequenceExtService,
	settings itExt.OrgSettingsReaderExtService,
) itStock.StockOperationTypeDomainService {
	return &StockOperationTypeDomainServiceImpl{CrudDomainService: base, sequences: sequences, settings: settings}
}

// StockOperationTypeDomainServiceImpl numbers the type's transfers from the organization's one
// inventory.transfer definition, on a counter of the type's own: the type is the scope object, so
// each type counts from 1. Creating or deleting a type registers nothing; the counter appears on
// the type's first transfer.
type StockOperationTypeDomainServiceImpl struct {
	dynamicresource.CrudDomainService
	sequences itExt.SequenceExtService
	settings  itExt.OrgSettingsReaderExtService
}

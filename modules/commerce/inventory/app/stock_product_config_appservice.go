package app

import (
	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// NewStockProductConfigApplicationService is handed the composable default by the stock_product_config onion.
func NewStockProductConfigApplicationService(base dynamicresource.CrudApplicationService) itStock.StockProductConfigApplicationService {
	return &StockProductConfigApplicationServiceImpl{CrudApplicationService: base}
}

// StockProductConfigApplicationServiceImpl is the authorized CRUD of the resource.
type StockProductConfigApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
}

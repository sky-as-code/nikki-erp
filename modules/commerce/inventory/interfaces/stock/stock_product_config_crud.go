package stock

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// StockProductConfigRepository reads and writes stock_product_config rows.
type StockProductConfigRepository interface {
	dynamicresource.CrudRepository
}

// StockProductConfigDomainService is the CRUD of the resource plus its own rules.
type StockProductConfigDomainService interface {
	dynamicresource.CrudDomainService
}

// StockProductConfigApplicationService is the authorized surface the REST handler serves.
type StockProductConfigApplicationService interface {
	dynamicresource.CrudApplicationService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateStockProductConfigCommand      = dynamicresource.CreateCommand
	UpdateStockProductConfigCommand      = dynamicresource.UpdateCommand
	DeleteStockProductConfigCommand      = dynamicresource.DeleteCommand
	SetStockProductConfigArchivedCommand = dynamicresource.SetArchivedCommand
	GetStockProductConfigByIdQuery       = dynamicresource.GetByIdQuery
	SearchStockProductConfigsQuery       = dynamicresource.SearchQuery
	StockProductConfigExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateStockProductConfigResult      = dynamicresource.CreateResult
	UpdateStockProductConfigResult      = dynamicresource.MutateResult
	DeleteStockProductConfigResult      = dynamicresource.MutateResult
	SetStockProductConfigArchivedResult = dynamicresource.MutateResult
	GetStockProductConfigByIdResult     = dynamicresource.GetOneResult
	SearchStockProductConfigsResult     = dynamicresource.SearchResult
	StockProductConfigExistsResult      = dynamicresource.ExistsResult
)

package stock

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// StockProductConfigRepository reads and writes stock_product_config rows.
type StockProductConfigRepository interface {
	composable.CrudRepository
}

// StockProductConfigDomainService is the CRUD of the resource plus its own rules.
type StockProductConfigDomainService interface {
	composable.CrudDomainService
}

// StockProductConfigApplicationService is the authorized surface the REST handler serves.
type StockProductConfigApplicationService interface {
	composable.CrudApplicationService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateStockProductConfigCommand      = composable.CreateCommand
	UpdateStockProductConfigCommand      = composable.UpdateCommand
	DeleteStockProductConfigCommand      = composable.DeleteCommand
	SetStockProductConfigArchivedCommand = composable.SetArchivedCommand
	GetStockProductConfigByIdQuery       = composable.GetByIdQuery
	SearchStockProductConfigsQuery       = composable.SearchQuery
	StockProductConfigExistsQuery        = composable.ExistsQuery
)

type (
	CreateStockProductConfigResult      = composable.CreateResult
	UpdateStockProductConfigResult      = composable.MutateResult
	DeleteStockProductConfigResult      = composable.MutateResult
	SetStockProductConfigArchivedResult = composable.MutateResult
	GetStockProductConfigByIdResult     = composable.GetOneResult
	SearchStockProductConfigsResult     = composable.SearchResult
	StockProductConfigExistsResult      = composable.ExistsResult
)

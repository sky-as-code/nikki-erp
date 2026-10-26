package stock

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// StockQuantRepository reads and writes stock_quant rows.
type StockQuantRepository interface {
	composable.CrudRepository
}

// StockQuantDomainService is the CRUD of the resource plus its own rules.
type StockQuantDomainService interface {
	composable.CrudDomainService
}

// StockQuantApplicationService is the authorized surface the REST handler serves.
type StockQuantApplicationService interface {
	composable.CrudApplicationService
	StockQuantActionService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateStockQuantCommand      = composable.CreateCommand
	UpdateStockQuantCommand      = composable.UpdateCommand
	DeleteStockQuantCommand      = composable.DeleteCommand
	SetStockQuantArchivedCommand = composable.SetArchivedCommand
	GetStockQuantByIdQuery       = composable.GetByIdQuery
	SearchStockQuantsQuery       = composable.SearchQuery
	StockQuantExistsQuery        = composable.ExistsQuery
)

type (
	CreateStockQuantResult      = composable.CreateResult
	UpdateStockQuantResult      = composable.MutateResult
	DeleteStockQuantResult      = composable.MutateResult
	SetStockQuantArchivedResult = composable.MutateResult
	GetStockQuantByIdResult     = composable.GetOneResult
	SearchStockQuantsResult     = composable.SearchResult
	StockQuantExistsResult      = composable.ExistsResult
)

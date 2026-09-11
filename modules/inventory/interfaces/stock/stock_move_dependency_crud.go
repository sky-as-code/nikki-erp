package stock

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// StockMoveDependencyRepository reads and writes stock_move_dependency rows.
type StockMoveDependencyRepository interface {
	composable.CrudRepository
}

// StockMoveDependencyDomainService is the CRUD of the resource plus its own rules.
type StockMoveDependencyDomainService interface {
	composable.CrudDomainService
}

// StockMoveDependencyApplicationService is the authorized surface the REST handler serves.
type StockMoveDependencyApplicationService interface {
	composable.CrudApplicationService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateStockMoveDependencyCommand      = composable.CreateCommand
	UpdateStockMoveDependencyCommand      = composable.UpdateCommand
	DeleteStockMoveDependencyCommand      = composable.DeleteCommand
	SetStockMoveDependencyArchivedCommand = composable.SetArchivedCommand
	GetStockMoveDependencyByIdQuery       = composable.GetByIdQuery
	SearchStockMoveDependenciesQuery      = composable.SearchQuery
	StockMoveDependencyExistsQuery        = composable.ExistsQuery
)

type (
	CreateStockMoveDependencyResult      = composable.CreateResult
	UpdateStockMoveDependencyResult      = composable.MutateResult
	DeleteStockMoveDependencyResult      = composable.MutateResult
	SetStockMoveDependencyArchivedResult = composable.MutateResult
	GetStockMoveDependencyByIdResult     = composable.GetOneResult
	SearchStockMoveDependenciesResult    = composable.SearchResult
	StockMoveDependencyExistsResult      = composable.ExistsResult
)

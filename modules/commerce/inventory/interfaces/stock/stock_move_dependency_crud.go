package stock

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// StockMoveDependencyRepository reads and writes stock_move_dependency rows.
type StockMoveDependencyRepository interface {
	dynamicresource.CrudRepository
}

// StockMoveDependencyDomainService is the CRUD of the resource plus its own rules.
type StockMoveDependencyDomainService interface {
	dynamicresource.CrudDomainService
}

// StockMoveDependencyApplicationService is the authorized surface the REST handler serves.
type StockMoveDependencyApplicationService interface {
	dynamicresource.CrudApplicationService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateStockMoveDependencyCommand      = dynamicresource.CreateCommand
	UpdateStockMoveDependencyCommand      = dynamicresource.UpdateCommand
	DeleteStockMoveDependencyCommand      = dynamicresource.DeleteCommand
	SetStockMoveDependencyArchivedCommand = dynamicresource.SetArchivedCommand
	GetStockMoveDependencyByIdQuery       = dynamicresource.GetByIdQuery
	SearchStockMoveDependenciesQuery      = dynamicresource.SearchQuery
	StockMoveDependencyExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateStockMoveDependencyResult      = dynamicresource.CreateResult
	UpdateStockMoveDependencyResult      = dynamicresource.MutateResult
	DeleteStockMoveDependencyResult      = dynamicresource.MutateResult
	SetStockMoveDependencyArchivedResult = dynamicresource.MutateResult
	GetStockMoveDependencyByIdResult     = dynamicresource.GetOneResult
	SearchStockMoveDependenciesResult    = dynamicresource.SearchResult
	StockMoveDependencyExistsResult      = dynamicresource.ExistsResult
)

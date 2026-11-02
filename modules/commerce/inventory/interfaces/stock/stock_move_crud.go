package stock

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// StockMoveRepository reads and writes stock_move rows.
type StockMoveRepository interface {
	dynamicresource.CrudRepository
}

// StockMoveDomainService is the CRUD of the resource plus its own rules.
type StockMoveDomainService interface {
	dynamicresource.CrudDomainService
}

// StockMoveApplicationService is the authorized surface the REST handler serves.
type StockMoveApplicationService interface {
	dynamicresource.CrudApplicationService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateStockMoveCommand      = dynamicresource.CreateCommand
	UpdateStockMoveCommand      = dynamicresource.UpdateCommand
	DeleteStockMoveCommand      = dynamicresource.DeleteCommand
	SetStockMoveArchivedCommand = dynamicresource.SetArchivedCommand
	GetStockMoveByIdQuery       = dynamicresource.GetByIdQuery
	SearchStockMovesQuery       = dynamicresource.SearchQuery
	StockMoveExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateStockMoveResult      = dynamicresource.CreateResult
	UpdateStockMoveResult      = dynamicresource.MutateResult
	DeleteStockMoveResult      = dynamicresource.MutateResult
	SetStockMoveArchivedResult = dynamicresource.MutateResult
	GetStockMoveByIdResult     = dynamicresource.GetOneResult
	SearchStockMovesResult     = dynamicresource.SearchResult
	StockMoveExistsResult      = dynamicresource.ExistsResult
)

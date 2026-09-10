package stock

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// StockMoveRepository reads and writes stock_move rows.
type StockMoveRepository interface {
	composable.CrudRepository
}

// StockMoveDomainService is the CRUD of the resource plus its own rules.
type StockMoveDomainService interface {
	composable.CrudDomainService
}

// StockMoveApplicationService is the authorized surface the REST handler serves.
type StockMoveApplicationService interface {
	composable.CrudApplicationService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateStockMoveCommand      = composable.CreateCommand
	UpdateStockMoveCommand      = composable.UpdateCommand
	DeleteStockMoveCommand      = composable.DeleteCommand
	SetStockMoveArchivedCommand = composable.SetArchivedCommand
	GetStockMoveByIdQuery       = composable.GetByIdQuery
	SearchStockMovesQuery       = composable.SearchQuery
	StockMoveExistsQuery        = composable.ExistsQuery
)

type (
	CreateStockMoveResult      = composable.CreateResult
	UpdateStockMoveResult      = composable.MutateResult
	DeleteStockMoveResult      = composable.MutateResult
	SetStockMoveArchivedResult = composable.MutateResult
	GetStockMoveByIdResult     = composable.GetOneResult
	SearchStockMovesResult     = composable.SearchResult
	StockMoveExistsResult      = composable.ExistsResult
)

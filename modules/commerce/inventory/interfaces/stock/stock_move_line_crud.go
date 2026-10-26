package stock

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// StockMoveLineRepository reads and writes stock_move_line rows.
type StockMoveLineRepository interface {
	composable.CrudRepository
}

// StockMoveLineDomainService is the CRUD of the resource plus its own rules.
type StockMoveLineDomainService interface {
	composable.CrudDomainService
}

// StockMoveLineApplicationService is the authorized surface the REST handler serves.
type StockMoveLineApplicationService interface {
	composable.CrudApplicationService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateStockMoveLineCommand      = composable.CreateCommand
	UpdateStockMoveLineCommand      = composable.UpdateCommand
	DeleteStockMoveLineCommand      = composable.DeleteCommand
	SetStockMoveLineArchivedCommand = composable.SetArchivedCommand
	GetStockMoveLineByIdQuery       = composable.GetByIdQuery
	SearchStockMoveLinesQuery       = composable.SearchQuery
	StockMoveLineExistsQuery        = composable.ExistsQuery
)

type (
	CreateStockMoveLineResult      = composable.CreateResult
	UpdateStockMoveLineResult      = composable.MutateResult
	DeleteStockMoveLineResult      = composable.MutateResult
	SetStockMoveLineArchivedResult = composable.MutateResult
	GetStockMoveLineByIdResult     = composable.GetOneResult
	SearchStockMoveLinesResult     = composable.SearchResult
	StockMoveLineExistsResult      = composable.ExistsResult
)

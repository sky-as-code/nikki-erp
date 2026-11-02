package stock

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// StockMoveLineRepository reads and writes stock_move_line rows.
type StockMoveLineRepository interface {
	dynamicresource.CrudRepository
}

// StockMoveLineDomainService is the CRUD of the resource plus its own rules.
type StockMoveLineDomainService interface {
	dynamicresource.CrudDomainService
}

// StockMoveLineApplicationService is the authorized surface the REST handler serves.
type StockMoveLineApplicationService interface {
	dynamicresource.CrudApplicationService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateStockMoveLineCommand      = dynamicresource.CreateCommand
	UpdateStockMoveLineCommand      = dynamicresource.UpdateCommand
	DeleteStockMoveLineCommand      = dynamicresource.DeleteCommand
	SetStockMoveLineArchivedCommand = dynamicresource.SetArchivedCommand
	GetStockMoveLineByIdQuery       = dynamicresource.GetByIdQuery
	SearchStockMoveLinesQuery       = dynamicresource.SearchQuery
	StockMoveLineExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateStockMoveLineResult      = dynamicresource.CreateResult
	UpdateStockMoveLineResult      = dynamicresource.MutateResult
	DeleteStockMoveLineResult      = dynamicresource.MutateResult
	SetStockMoveLineArchivedResult = dynamicresource.MutateResult
	GetStockMoveLineByIdResult     = dynamicresource.GetOneResult
	SearchStockMoveLinesResult     = dynamicresource.SearchResult
	StockMoveLineExistsResult      = dynamicresource.ExistsResult
)

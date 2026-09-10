package stock

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// StockTransferRepository reads and writes stock_transfer rows.
type StockTransferRepository interface {
	composable.CrudRepository
}

// StockTransferDomainService is the CRUD of the resource plus its own rules.
type StockTransferDomainService interface {
	composable.CrudDomainService
}

// StockTransferApplicationService is the authorized surface the REST handler serves.
type StockTransferApplicationService interface {
	composable.CrudApplicationService
	StockTransferActionService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateStockTransferCommand      = composable.CreateCommand
	UpdateStockTransferCommand      = composable.UpdateCommand
	DeleteStockTransferCommand      = composable.DeleteCommand
	SetStockTransferArchivedCommand = composable.SetArchivedCommand
	GetStockTransferByIdQuery       = composable.GetByIdQuery
	SearchStockTransfersQuery       = composable.SearchQuery
	StockTransferExistsQuery        = composable.ExistsQuery
)

type (
	CreateStockTransferResult      = composable.CreateResult
	UpdateStockTransferResult      = composable.MutateResult
	DeleteStockTransferResult      = composable.MutateResult
	SetStockTransferArchivedResult = composable.MutateResult
	GetStockTransferByIdResult     = composable.GetOneResult
	SearchStockTransfersResult     = composable.SearchResult
	StockTransferExistsResult      = composable.ExistsResult
)

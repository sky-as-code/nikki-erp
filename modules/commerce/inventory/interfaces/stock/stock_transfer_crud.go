package stock

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// StockTransferRepository reads and writes stock_transfer rows.
type StockTransferRepository interface {
	dynamicresource.CrudRepository
}

// StockTransferDomainService is the CRUD of the resource plus its own rules.
type StockTransferDomainService interface {
	dynamicresource.CrudDomainService
}

// StockTransferApplicationService is the authorized surface the REST handler serves.
type StockTransferApplicationService interface {
	dynamicresource.CrudApplicationService
	StockTransferActionService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateStockTransferCommand      = dynamicresource.CreateCommand
	UpdateStockTransferCommand      = dynamicresource.UpdateCommand
	DeleteStockTransferCommand      = dynamicresource.DeleteCommand
	SetStockTransferArchivedCommand = dynamicresource.SetArchivedCommand
	GetStockTransferByIdQuery       = dynamicresource.GetByIdQuery
	SearchStockTransfersQuery       = dynamicresource.SearchQuery
	StockTransferExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateStockTransferResult      = dynamicresource.CreateResult
	UpdateStockTransferResult      = dynamicresource.MutateResult
	DeleteStockTransferResult      = dynamicresource.MutateResult
	SetStockTransferArchivedResult = dynamicresource.MutateResult
	GetStockTransferByIdResult     = dynamicresource.GetOneResult
	SearchStockTransfersResult     = dynamicresource.SearchResult
	StockTransferExistsResult      = dynamicresource.ExistsResult
)

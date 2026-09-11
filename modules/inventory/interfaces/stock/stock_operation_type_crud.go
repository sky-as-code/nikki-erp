package stock

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// StockOperationTypeRepository reads and writes stock_operation_type rows.
type StockOperationTypeRepository interface {
	composable.CrudRepository
}

// StockOperationTypeDomainService is the CRUD of the resource plus its own rules.
type StockOperationTypeDomainService interface {
	composable.CrudDomainService
}

// StockOperationTypeApplicationService is the authorized surface the REST handler serves.
type StockOperationTypeApplicationService interface {
	composable.CrudApplicationService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateStockOperationTypeCommand      = composable.CreateCommand
	UpdateStockOperationTypeCommand      = composable.UpdateCommand
	DeleteStockOperationTypeCommand      = composable.DeleteCommand
	SetStockOperationTypeArchivedCommand = composable.SetArchivedCommand
	GetStockOperationTypeByIdQuery       = composable.GetByIdQuery
	SearchStockOperationTypesQuery       = composable.SearchQuery
	StockOperationTypeExistsQuery        = composable.ExistsQuery
)

type (
	CreateStockOperationTypeResult      = composable.CreateResult
	UpdateStockOperationTypeResult      = composable.MutateResult
	DeleteStockOperationTypeResult      = composable.MutateResult
	SetStockOperationTypeArchivedResult = composable.MutateResult
	GetStockOperationTypeByIdResult     = composable.GetOneResult
	SearchStockOperationTypesResult     = composable.SearchResult
	StockOperationTypeExistsResult      = composable.ExistsResult
)

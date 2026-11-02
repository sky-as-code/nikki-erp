package stock

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// StockOperationTypeRepository reads and writes stock_operation_type rows.
type StockOperationTypeRepository interface {
	dynamicresource.CrudRepository
}

// StockOperationTypeDomainService is the CRUD of the resource plus its own rules.
type StockOperationTypeDomainService interface {
	dynamicresource.CrudDomainService
}

// StockOperationTypeApplicationService is the authorized surface the REST handler serves.
type StockOperationTypeApplicationService interface {
	dynamicresource.CrudApplicationService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateStockOperationTypeCommand      = dynamicresource.CreateCommand
	UpdateStockOperationTypeCommand      = dynamicresource.UpdateCommand
	DeleteStockOperationTypeCommand      = dynamicresource.DeleteCommand
	SetStockOperationTypeArchivedCommand = dynamicresource.SetArchivedCommand
	GetStockOperationTypeByIdQuery       = dynamicresource.GetByIdQuery
	SearchStockOperationTypesQuery       = dynamicresource.SearchQuery
	StockOperationTypeExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateStockOperationTypeResult      = dynamicresource.CreateResult
	UpdateStockOperationTypeResult      = dynamicresource.MutateResult
	DeleteStockOperationTypeResult      = dynamicresource.MutateResult
	SetStockOperationTypeArchivedResult = dynamicresource.MutateResult
	GetStockOperationTypeByIdResult     = dynamicresource.GetOneResult
	SearchStockOperationTypesResult     = dynamicresource.SearchResult
	StockOperationTypeExistsResult      = dynamicresource.ExistsResult
)

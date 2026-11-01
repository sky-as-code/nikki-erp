package stock

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// StockQuantRepository reads and writes stock_quant rows.
type StockQuantRepository interface {
	dynamicresource.CrudRepository
}

// StockQuantDomainService is the CRUD of the resource plus its own rules.
type StockQuantDomainService interface {
	dynamicresource.CrudDomainService
}

// StockQuantApplicationService is the authorized surface the REST handler serves.
type StockQuantApplicationService interface {
	dynamicresource.CrudApplicationService
	StockQuantActionService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateStockQuantCommand      = dynamicresource.CreateCommand
	UpdateStockQuantCommand      = dynamicresource.UpdateCommand
	DeleteStockQuantCommand      = dynamicresource.DeleteCommand
	SetStockQuantArchivedCommand = dynamicresource.SetArchivedCommand
	GetStockQuantByIdQuery       = dynamicresource.GetByIdQuery
	SearchStockQuantsQuery       = dynamicresource.SearchQuery
	StockQuantExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateStockQuantResult      = dynamicresource.CreateResult
	UpdateStockQuantResult      = dynamicresource.MutateResult
	DeleteStockQuantResult      = dynamicresource.MutateResult
	SetStockQuantArchivedResult = dynamicresource.MutateResult
	GetStockQuantByIdResult     = dynamicresource.GetOneResult
	SearchStockQuantsResult     = dynamicresource.SearchResult
	StockQuantExistsResult      = dynamicresource.ExistsResult
)

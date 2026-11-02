package stock

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// StockScrapRepository reads and writes stock_scrap rows.
type StockScrapRepository interface {
	dynamicresource.CrudRepository
}

// StockScrapDomainService is the CRUD of the resource plus its own rules.
type StockScrapDomainService interface {
	dynamicresource.CrudDomainService
}

// StockScrapApplicationService is the authorized surface the REST handler serves.
type StockScrapApplicationService interface {
	dynamicresource.CrudApplicationService
	StockScrapActionService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateStockScrapCommand      = dynamicresource.CreateCommand
	UpdateStockScrapCommand      = dynamicresource.UpdateCommand
	DeleteStockScrapCommand      = dynamicresource.DeleteCommand
	SetStockScrapArchivedCommand = dynamicresource.SetArchivedCommand
	GetStockScrapByIdQuery       = dynamicresource.GetByIdQuery
	SearchStockScrapsQuery       = dynamicresource.SearchQuery
	StockScrapExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateStockScrapResult      = dynamicresource.CreateResult
	UpdateStockScrapResult      = dynamicresource.MutateResult
	DeleteStockScrapResult      = dynamicresource.MutateResult
	SetStockScrapArchivedResult = dynamicresource.MutateResult
	GetStockScrapByIdResult     = dynamicresource.GetOneResult
	SearchStockScrapsResult     = dynamicresource.SearchResult
	StockScrapExistsResult      = dynamicresource.ExistsResult
)

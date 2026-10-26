package stock

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// StockScrapRepository reads and writes stock_scrap rows.
type StockScrapRepository interface {
	composable.CrudRepository
}

// StockScrapDomainService is the CRUD of the resource plus its own rules.
type StockScrapDomainService interface {
	composable.CrudDomainService
}

// StockScrapApplicationService is the authorized surface the REST handler serves.
type StockScrapApplicationService interface {
	composable.CrudApplicationService
	StockScrapActionService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateStockScrapCommand      = composable.CreateCommand
	UpdateStockScrapCommand      = composable.UpdateCommand
	DeleteStockScrapCommand      = composable.DeleteCommand
	SetStockScrapArchivedCommand = composable.SetArchivedCommand
	GetStockScrapByIdQuery       = composable.GetByIdQuery
	SearchStockScrapsQuery       = composable.SearchQuery
	StockScrapExistsQuery        = composable.ExistsQuery
)

type (
	CreateStockScrapResult      = composable.CreateResult
	UpdateStockScrapResult      = composable.MutateResult
	DeleteStockScrapResult      = composable.MutateResult
	SetStockScrapArchivedResult = composable.MutateResult
	GetStockScrapByIdResult     = composable.GetOneResult
	SearchStockScrapsResult     = composable.SearchResult
	StockScrapExistsResult      = composable.ExistsResult
)

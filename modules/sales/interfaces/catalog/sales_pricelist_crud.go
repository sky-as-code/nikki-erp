package catalog

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// SalesPricelistRepository reads and writes pricelists.
type SalesPricelistRepository interface {
	composable.CrudRepository
}

// SalesPricelistDomainService adds the rules the schema cannot express: at most one default per
// organization among non-archived rows, and the refusal to restate a pricelist's currency once
// prices are quoted in it.
type SalesPricelistDomainService interface {
	composable.CrudDomainService

	// SetDefault promotes one pricelist and demotes the incumbent in the same transaction, so no
	// window exists in which an organization has two defaults or none.
	SetDefault(ctx corectx.Context, pricelistId string) (*dyn.OpResult[dyn.MutateResultData], error)
}

// SalesPricelistApplicationService is the authorized surface the REST handler serves.
type SalesPricelistApplicationService interface {
	composable.CrudApplicationService

	SetDefault(ctx corectx.Context, cmd SetDefaultPricelistCommand) (*SetDefaultPricelistResult, error)
}

// SalesPricelistItemRepository reads and writes the priced lines of a pricelist.
type SalesPricelistItemRepository interface {
	composable.CrudRepository
}

// SalesPricelistItemDomainService carries the rules conditional on another field, which the schema
// cannot state: which target column is required depends on applies_to, which price fields are
// required depends on calculation_method, and whether a base pricelist is acceptable depends on
// the whole derivation graph.
type SalesPricelistItemDomainService interface {
	composable.CrudDomainService
}

// SalesPricelistItemApplicationService is the authorized surface the REST handler serves.
type SalesPricelistItemApplicationService interface {
	composable.CrudApplicationService
}

// SetDefaultPricelistCommand names the pricelist to promote. It is its own operation rather than
// an ordinary update because promoting one list must demote the other in the same breath.
type (
	SetDefaultPricelistCommand = composable.UpdateCommand
	SetDefaultPricelistResult  = composable.MutateResult
)

type (
	CreateSalesPricelistCommand      = composable.CreateCommand
	UpdateSalesPricelistCommand      = composable.UpdateCommand
	DeleteSalesPricelistCommand      = composable.DeleteCommand
	SetSalesPricelistArchivedCommand = composable.SetArchivedCommand
	GetSalesPricelistByIdQuery       = composable.GetByIdQuery
	SearchSalesPricelistsQuery       = composable.SearchQuery
	SalesPricelistExistsQuery        = composable.ExistsQuery
)

type (
	CreateSalesPricelistResult      = composable.CreateResult
	UpdateSalesPricelistResult      = composable.MutateResult
	DeleteSalesPricelistResult      = composable.MutateResult
	SetSalesPricelistArchivedResult = composable.MutateResult
	GetSalesPricelistByIdResult     = composable.GetOneResult
	SearchSalesPricelistsResult     = composable.SearchResult
	SalesPricelistExistsResult      = composable.ExistsResult
)

type (
	CreateSalesPricelistItemCommand      = composable.CreateCommand
	UpdateSalesPricelistItemCommand      = composable.UpdateCommand
	DeleteSalesPricelistItemCommand      = composable.DeleteCommand
	SetSalesPricelistItemArchivedCommand = composable.SetArchivedCommand
	GetSalesPricelistItemByIdQuery       = composable.GetByIdQuery
	SearchSalesPricelistItemsQuery       = composable.SearchQuery
	SalesPricelistItemExistsQuery        = composable.ExistsQuery
)

type (
	CreateSalesPricelistItemResult      = composable.CreateResult
	UpdateSalesPricelistItemResult      = composable.MutateResult
	DeleteSalesPricelistItemResult      = composable.MutateResult
	SetSalesPricelistItemArchivedResult = composable.MutateResult
	GetSalesPricelistItemByIdResult     = composable.GetOneResult
	SearchSalesPricelistItemsResult     = composable.SearchResult
	SalesPricelistItemExistsResult      = composable.ExistsResult
)

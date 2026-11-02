package catalog

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// SalesPricelistRepository reads and writes pricelists.
type SalesPricelistRepository interface {
	dynamicresource.CrudRepository
}

// SalesPricelistDomainService adds the rules the schema cannot express: at most one default per
// organization among non-archived rows, and the refusal to restate a pricelist's currency once
// prices are quoted in it.
type SalesPricelistDomainService interface {
	dynamicresource.CrudDomainService

	// SetDefault promotes one pricelist and demotes the incumbent in the same transaction, so no
	// window exists in which an organization has two defaults or none.
	SetDefault(ctx corectx.Context, pricelistId string) (*dyn.OpResult[dyn.MutateResultData], error)
}

// SalesPricelistApplicationService is the authorized surface the REST handler serves.
type SalesPricelistApplicationService interface {
	dynamicresource.CrudApplicationService

	SetDefault(ctx corectx.Context, cmd SetDefaultPricelistCommand) (*SetDefaultPricelistResult, error)
}

// SalesPricelistItemRepository reads and writes the priced lines of a pricelist.
type SalesPricelistItemRepository interface {
	dynamicresource.CrudRepository
}

// SalesPricelistItemDomainService carries the rules conditional on another field, which the schema
// cannot state: which target column is required depends on applies_to, which price fields are
// required depends on calculation_method, and whether a base pricelist is acceptable depends on
// the whole derivation graph.
type SalesPricelistItemDomainService interface {
	dynamicresource.CrudDomainService
}

// SalesPricelistItemApplicationService is the authorized surface the REST handler serves.
type SalesPricelistItemApplicationService interface {
	dynamicresource.CrudApplicationService
}

// SetDefaultPricelistCommand names the pricelist to promote. It is its own operation rather than
// an ordinary update because promoting one list must demote the other in the same breath.
type (
	SetDefaultPricelistCommand = dynamicresource.UpdateCommand
	SetDefaultPricelistResult  = dynamicresource.MutateResult
)

type (
	CreateSalesPricelistCommand      = dynamicresource.CreateCommand
	UpdateSalesPricelistCommand      = dynamicresource.UpdateCommand
	DeleteSalesPricelistCommand      = dynamicresource.DeleteCommand
	SetSalesPricelistArchivedCommand = dynamicresource.SetArchivedCommand
	GetSalesPricelistByIdQuery       = dynamicresource.GetByIdQuery
	SearchSalesPricelistsQuery       = dynamicresource.SearchQuery
	SalesPricelistExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateSalesPricelistResult      = dynamicresource.CreateResult
	UpdateSalesPricelistResult      = dynamicresource.MutateResult
	DeleteSalesPricelistResult      = dynamicresource.MutateResult
	SetSalesPricelistArchivedResult = dynamicresource.MutateResult
	GetSalesPricelistByIdResult     = dynamicresource.GetOneResult
	SearchSalesPricelistsResult     = dynamicresource.SearchResult
	SalesPricelistExistsResult      = dynamicresource.ExistsResult
)

type (
	CreateSalesPricelistItemCommand      = dynamicresource.CreateCommand
	UpdateSalesPricelistItemCommand      = dynamicresource.UpdateCommand
	DeleteSalesPricelistItemCommand      = dynamicresource.DeleteCommand
	SetSalesPricelistItemArchivedCommand = dynamicresource.SetArchivedCommand
	GetSalesPricelistItemByIdQuery       = dynamicresource.GetByIdQuery
	SearchSalesPricelistItemsQuery       = dynamicresource.SearchQuery
	SalesPricelistItemExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateSalesPricelistItemResult      = dynamicresource.CreateResult
	UpdateSalesPricelistItemResult      = dynamicresource.MutateResult
	DeleteSalesPricelistItemResult      = dynamicresource.MutateResult
	SetSalesPricelistItemArchivedResult = dynamicresource.MutateResult
	GetSalesPricelistItemByIdResult     = dynamicresource.GetOneResult
	SearchSalesPricelistItemsResult     = dynamicresource.SearchResult
	SalesPricelistItemExistsResult      = dynamicresource.ExistsResult
)

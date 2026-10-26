package product

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// ProductVariantRepository reads and writes product_variant rows.
type ProductVariantRepository interface {
	composable.CrudRepository
}

// ProductVariantApplicationService is the authorized surface the REST handler serves.
type ProductVariantApplicationService interface {
	composable.CrudApplicationService
	ProductVariantActionService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateProductVariantCommand      = composable.CreateCommand
	UpdateProductVariantCommand      = composable.UpdateCommand
	DeleteProductVariantCommand      = composable.DeleteCommand
	SetProductVariantArchivedCommand = composable.SetArchivedCommand
	GetProductVariantByIdQuery       = composable.GetByIdQuery
	SearchProductVariantRowsQuery    = composable.SearchQuery
	ProductVariantExistsQuery        = composable.ExistsQuery
)

type (
	CreateProductVariantResult      = composable.CreateResult
	UpdateProductVariantResult      = composable.MutateResult
	DeleteProductVariantResult      = composable.MutateResult
	SetProductVariantArchivedResult = composable.MutateResult
	GetProductVariantByIdResult     = composable.GetOneResult
	SearchProductVariantRowsResult  = composable.SearchResult
	ProductVariantExistsResult      = composable.ExistsResult
)

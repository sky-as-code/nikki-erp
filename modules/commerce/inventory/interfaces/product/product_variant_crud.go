package product

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// ProductVariantRepository reads and writes product_variant rows.
type ProductVariantRepository interface {
	dynamicresource.CrudRepository
}

// ProductVariantApplicationService is the authorized surface the REST handler serves.
type ProductVariantApplicationService interface {
	dynamicresource.CrudApplicationService
	ProductVariantActionService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateProductVariantCommand      = dynamicresource.CreateCommand
	UpdateProductVariantCommand      = dynamicresource.UpdateCommand
	DeleteProductVariantCommand      = dynamicresource.DeleteCommand
	SetProductVariantArchivedCommand = dynamicresource.SetArchivedCommand
	GetProductVariantByIdQuery       = dynamicresource.GetByIdQuery
	SearchProductVariantRowsQuery    = dynamicresource.SearchQuery
	ProductVariantExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateProductVariantResult      = dynamicresource.CreateResult
	UpdateProductVariantResult      = dynamicresource.MutateResult
	DeleteProductVariantResult      = dynamicresource.MutateResult
	SetProductVariantArchivedResult = dynamicresource.MutateResult
	GetProductVariantByIdResult     = dynamicresource.GetOneResult
	SearchProductVariantRowsResult  = dynamicresource.SearchResult
	ProductVariantExistsResult      = dynamicresource.ExistsResult
)

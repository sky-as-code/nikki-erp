package product

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// The Product Variant port other modules bind to.
//
// A variant is what is physically stocked and sold, so a consumer that stocks shelves works in
// variants rather than templates. Reading one now also yields its template's display fields
// through the template_* virtual fields, so a caller no longer needs a second read to label what
// it is showing.
//
// Inventory deliberately no longer uses the word "catalog": this is the Product Variant service,
// named for the single resource it owns.

type SearchProductVariantsQuery = dyn.SearchQuery
type SearchProductVariantsResultData = dyn.PagedResultData[models.ProductVariant]
type SearchProductVariantsResult = dyn.OpResult[SearchProductVariantsResultData]

type GetProductVariantQuery = dyn.GetOneQuery
type GetProductVariantResult = dyn.OpResult[models.ProductVariant]

type ProductVariantsExistQuery = dyn.ExistsQuery
type ProductVariantsExistResult = dyn.OpResult[dyn.ExistsResultData]

// ProductVariantDomainService reads product variants on behalf of another module.
//
// Kept deliberately small: it covers what a consumer needs to resolve and validate the variants
// it references, and nothing that would let it manage master data. Writes stay with Inventory.
type ProductVariantDomainService interface {
	// SearchProductVariants finds variants matching a search graph.
	//
	// Requesting a template_* field fills it for the whole page in one batched read, so a listing
	// costs two queries however many rows it returns. Filtering or sorting on one works too: the
	// service rewrites it to the underlying template edge path.
	SearchProductVariants(
		ctx corectx.Context, query SearchProductVariantsQuery,
	) (*SearchProductVariantsResult, error)

	// GetProductVariant reads one variant by id, including its template_* fields.
	GetProductVariant(ctx corectx.Context, query GetProductVariantQuery) (*GetProductVariantResult, error)

	// ProductVariantsExist reports which of the given ids exist, so a caller can validate a batch
	// of references without reading every record.
	ProductVariantsExist(
		ctx corectx.Context, query ProductVariantsExistQuery,
	) (*ProductVariantsExistResult, error)
}

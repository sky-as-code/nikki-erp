package product

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// The Product Variant port other modules bind to. A variant is what is physically stocked and
// sold; reading one also yields its template's display fields through the template_* virtual
// fields, so no second read is needed to label it.

type SearchProductVariantsQuery = dyn.SearchQuery
type SearchProductVariantsResultData = dyn.PagedResultData[models.ProductVariant]
type SearchProductVariantsResult = dyn.OpResult[SearchProductVariantsResultData]

type GetProductVariantQuery = dyn.GetOneQuery
type GetProductVariantResult = dyn.OpResult[models.ProductVariant]

type ProductVariantsExistQuery = dyn.ExistsQuery
type ProductVariantsExistResult = dyn.OpResult[dyn.ExistsResultData]

// ProductVariantDomainService reads product variants on behalf of another module. Deliberately
// read-only: a consumer can resolve and validate references but not manage master data.
type ProductVariantDomainService interface {
	// SearchProductVariants finds variants matching a search graph. A requested template_* field
	// is filled for the whole page in one batched read, so a listing costs two queries whatever
	// its row count; filtering and sorting on one are rewritten to the template edge path.
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

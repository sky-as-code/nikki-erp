package product

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// The read-only ports for the resources around a variant.
//
// These used to sit on one "catalog" service alongside the variant reads. They are separated
// because they are not variant operations: a consumer grouping products by the thing a customer
// recognises reads templates, and one displaying a filing category reads categories. Splitting
// them keeps each port named for the resource it actually owns.

type SearchTemplatesQuery = dyn.SearchQuery
type SearchTemplatesResultData = dyn.PagedResultData[models.ProductTemplate]
type SearchTemplatesResult = dyn.OpResult[SearchTemplatesResultData]

type SearchCategoriesQuery = dyn.SearchQuery
type SearchCategoriesResultData = dyn.PagedResultData[models.ProductCategory]
type SearchCategoriesResult = dyn.OpResult[SearchCategoriesResultData]

// ProductTemplateReadService finds templates matching a search graph, for the consumer that
// groups variants by the product a customer recognises.
type ProductTemplateReadService interface {
	SearchTemplates(ctx corectx.Context, query SearchTemplatesQuery) (*SearchTemplatesResult, error)
}

// ProductCategoryReadService resolves categories, for a consumer that displays the category a
// product is filed under. A template belongs to exactly one.
type ProductCategoryReadService interface {
	SearchCategories(ctx corectx.Context, query SearchCategoriesQuery) (*SearchCategoriesResult, error)
}

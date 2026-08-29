package product

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// Read-only ports for the resources around a variant, one port per resource owned.

type SearchTemplatesQuery = dyn.SearchQuery
type SearchTemplatesResultData = dyn.PagedResultData[models.ProductTemplate]
type SearchTemplatesResult = dyn.OpResult[SearchTemplatesResultData]

type SearchCategoriesQuery = dyn.SearchQuery
type SearchCategoriesResultData = dyn.PagedResultData[models.ProductCategory]
type SearchCategoriesResult = dyn.OpResult[SearchCategoriesResultData]

// ProductTemplateReadService finds templates matching a search graph.
type ProductTemplateReadService interface {
	SearchTemplates(ctx corectx.Context, query SearchTemplatesQuery) (*SearchTemplatesResult, error)
}

// ProductCategoryReadService resolves categories. A template belongs to exactly one.
type ProductCategoryReadService interface {
	SearchCategories(ctx corectx.Context, query SearchCategoriesQuery) (*SearchCategoriesResult, error)
}

package product

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// ProductCategoryRepository reads and writes product_category rows.
type ProductCategoryRepository interface {
	dynamicresource.CrudRepository
}

// ProductCategoryDomainService is the CRUD of the resource plus its own rules.
type ProductCategoryDomainService interface {
	dynamicresource.CrudDomainService

	// CreateRoot creates an org's root category; the org lifecycle calls it once per org.
	CreateRoot(ctx corectx.Context, cmd CreateProductCategoryCommand) (*CreateProductCategoryResult, error)
}

// ProductCategoryApplicationService is the authorized surface the REST handler serves.
type ProductCategoryApplicationService interface {
	dynamicresource.CrudApplicationService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateProductCategoryCommand      = dynamicresource.CreateCommand
	UpdateProductCategoryCommand      = dynamicresource.UpdateCommand
	DeleteProductCategoryCommand      = dynamicresource.DeleteCommand
	SetProductCategoryArchivedCommand = dynamicresource.SetArchivedCommand
	GetProductCategoryByIdQuery       = dynamicresource.GetByIdQuery
	SearchProductCategoriesQuery      = dynamicresource.SearchQuery
	ProductCategoryExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateProductCategoryResult      = dynamicresource.CreateResult
	UpdateProductCategoryResult      = dynamicresource.MutateResult
	DeleteProductCategoryResult      = dynamicresource.MutateResult
	SetProductCategoryArchivedResult = dynamicresource.MutateResult
	GetProductCategoryByIdResult     = dynamicresource.GetOneResult
	SearchProductCategoriesResult    = dynamicresource.SearchResult
	ProductCategoryExistsResult      = dynamicresource.ExistsResult
)

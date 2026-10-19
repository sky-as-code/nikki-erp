package product

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// ProductCategoryRepository reads and writes product_category rows.
type ProductCategoryRepository interface {
	composable.CrudRepository
}

// ProductCategoryDomainService is the CRUD of the resource plus its own rules.
type ProductCategoryDomainService interface {
	composable.CrudDomainService

	// CreateRoot creates an org's root category; the org lifecycle calls it once per org.
	CreateRoot(ctx corectx.Context, cmd CreateProductCategoryCommand) (*CreateProductCategoryResult, error)
}

// ProductCategoryApplicationService is the authorized surface the REST handler serves.
type ProductCategoryApplicationService interface {
	composable.CrudApplicationService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateProductCategoryCommand      = composable.CreateCommand
	UpdateProductCategoryCommand      = composable.UpdateCommand
	DeleteProductCategoryCommand      = composable.DeleteCommand
	SetProductCategoryArchivedCommand = composable.SetArchivedCommand
	GetProductCategoryByIdQuery       = composable.GetByIdQuery
	SearchProductCategoriesQuery      = composable.SearchQuery
	ProductCategoryExistsQuery        = composable.ExistsQuery
)

type (
	CreateProductCategoryResult      = composable.CreateResult
	UpdateProductCategoryResult      = composable.MutateResult
	DeleteProductCategoryResult      = composable.MutateResult
	SetProductCategoryArchivedResult = composable.MutateResult
	GetProductCategoryByIdResult     = composable.GetOneResult
	SearchProductCategoriesResult    = composable.SearchResult
	ProductCategoryExistsResult      = composable.ExistsResult
)

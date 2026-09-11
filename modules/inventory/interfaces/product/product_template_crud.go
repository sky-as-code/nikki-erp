package product

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// ProductTemplateRepository reads and writes product_template rows.
type ProductTemplateRepository interface {
	composable.CrudRepository
}

// ProductTemplateDomainService is the Products capability under the layer name every resource
// declares; ProductService is the name consumers already know it by.
type ProductTemplateDomainService = ProductService

// ProductTemplateApplicationService is the authorized surface the REST handler serves.
type ProductTemplateApplicationService interface {
	composable.CrudApplicationService
	ProductTemplateActionService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateProductTemplateCommand      = composable.CreateCommand
	UpdateProductTemplateCommand      = composable.UpdateCommand
	DeleteProductTemplateCommand      = composable.DeleteCommand
	SetProductTemplateArchivedCommand = composable.SetArchivedCommand
	GetProductTemplateByIdQuery       = composable.GetByIdQuery
	SearchProductTemplatesQuery       = composable.SearchQuery
	ProductTemplateExistsQuery        = composable.ExistsQuery
)

type (
	CreateProductTemplateResult      = composable.CreateResult
	UpdateProductTemplateResult      = composable.MutateResult
	DeleteProductTemplateResult      = composable.MutateResult
	SetProductTemplateArchivedResult = composable.MutateResult
	GetProductTemplateByIdResult     = composable.GetOneResult
	SearchProductTemplatesResult     = composable.SearchResult
	ProductTemplateExistsResult      = composable.ExistsResult
)

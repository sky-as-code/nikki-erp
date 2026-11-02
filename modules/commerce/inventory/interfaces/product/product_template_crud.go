package product

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// ProductTemplateRepository reads and writes product_template rows.
type ProductTemplateRepository interface {
	dynamicresource.CrudRepository
}

// ProductTemplateDomainService is the Products capability under the layer name every resource
// declares; ProductService is the name consumers already know it by.
type ProductTemplateDomainService = ProductService

// ProductTemplateApplicationService is the authorized surface the REST handler serves.
type ProductTemplateApplicationService interface {
	dynamicresource.CrudApplicationService
	ProductTemplateActionService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateProductTemplateCommand      = dynamicresource.CreateCommand
	UpdateProductTemplateCommand      = dynamicresource.UpdateCommand
	DeleteProductTemplateCommand      = dynamicresource.DeleteCommand
	SetProductTemplateArchivedCommand = dynamicresource.SetArchivedCommand
	GetProductTemplateByIdQuery       = dynamicresource.GetByIdQuery
	SearchProductTemplatesQuery       = dynamicresource.SearchQuery
	ProductTemplateExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateProductTemplateResult      = dynamicresource.CreateResult
	UpdateProductTemplateResult      = dynamicresource.MutateResult
	DeleteProductTemplateResult      = dynamicresource.MutateResult
	SetProductTemplateArchivedResult = dynamicresource.MutateResult
	GetProductTemplateByIdResult     = dynamicresource.GetOneResult
	SearchProductTemplatesResult     = dynamicresource.SearchResult
	ProductTemplateExistsResult      = dynamicresource.ExistsResult
)

package product

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// ProductTemplateAttributeRepository reads and writes product_template_attribute rows.
type ProductTemplateAttributeRepository interface {
	dynamicresource.CrudRepository
}

// ProductTemplateAttributeDomainService is the CRUD of the resource plus its own rules.
type ProductTemplateAttributeDomainService interface {
	dynamicresource.CrudDomainService
}

// ProductTemplateAttributeApplicationService is the authorized surface the REST handler serves.
type ProductTemplateAttributeApplicationService interface {
	dynamicresource.CrudApplicationService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateProductTemplateAttributeCommand      = dynamicresource.CreateCommand
	UpdateProductTemplateAttributeCommand      = dynamicresource.UpdateCommand
	DeleteProductTemplateAttributeCommand      = dynamicresource.DeleteCommand
	SetProductTemplateAttributeArchivedCommand = dynamicresource.SetArchivedCommand
	GetProductTemplateAttributeByIdQuery       = dynamicresource.GetByIdQuery
	SearchProductTemplateAttributesQuery       = dynamicresource.SearchQuery
	ProductTemplateAttributeExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateProductTemplateAttributeResult      = dynamicresource.CreateResult
	UpdateProductTemplateAttributeResult      = dynamicresource.MutateResult
	DeleteProductTemplateAttributeResult      = dynamicresource.MutateResult
	SetProductTemplateAttributeArchivedResult = dynamicresource.MutateResult
	GetProductTemplateAttributeByIdResult     = dynamicresource.GetOneResult
	SearchProductTemplateAttributesResult     = dynamicresource.SearchResult
	ProductTemplateAttributeExistsResult      = dynamicresource.ExistsResult
)

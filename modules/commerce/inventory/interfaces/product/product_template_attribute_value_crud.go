package product

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// ProductTemplateAttributeValueRepository reads and writes product_template_attribute_value rows.
type ProductTemplateAttributeValueRepository interface {
	dynamicresource.CrudRepository
}

// ProductTemplateAttributeValueDomainService is the CRUD of the resource plus its own rules.
type ProductTemplateAttributeValueDomainService interface {
	dynamicresource.CrudDomainService
}

// ProductTemplateAttributeValueApplicationService is the authorized surface the REST handler serves.
type ProductTemplateAttributeValueApplicationService interface {
	dynamicresource.CrudApplicationService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateProductTemplateAttributeValueCommand      = dynamicresource.CreateCommand
	UpdateProductTemplateAttributeValueCommand      = dynamicresource.UpdateCommand
	DeleteProductTemplateAttributeValueCommand      = dynamicresource.DeleteCommand
	SetProductTemplateAttributeValueArchivedCommand = dynamicresource.SetArchivedCommand
	GetProductTemplateAttributeValueByIdQuery       = dynamicresource.GetByIdQuery
	SearchProductTemplateAttributeValuesQuery       = dynamicresource.SearchQuery
	ProductTemplateAttributeValueExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateProductTemplateAttributeValueResult      = dynamicresource.CreateResult
	UpdateProductTemplateAttributeValueResult      = dynamicresource.MutateResult
	DeleteProductTemplateAttributeValueResult      = dynamicresource.MutateResult
	SetProductTemplateAttributeValueArchivedResult = dynamicresource.MutateResult
	GetProductTemplateAttributeValueByIdResult     = dynamicresource.GetOneResult
	SearchProductTemplateAttributeValuesResult     = dynamicresource.SearchResult
	ProductTemplateAttributeValueExistsResult      = dynamicresource.ExistsResult
)

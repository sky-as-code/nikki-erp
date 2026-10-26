package product

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// ProductTemplateAttributeValueRepository reads and writes product_template_attribute_value rows.
type ProductTemplateAttributeValueRepository interface {
	composable.CrudRepository
}

// ProductTemplateAttributeValueDomainService is the CRUD of the resource plus its own rules.
type ProductTemplateAttributeValueDomainService interface {
	composable.CrudDomainService
}

// ProductTemplateAttributeValueApplicationService is the authorized surface the REST handler serves.
type ProductTemplateAttributeValueApplicationService interface {
	composable.CrudApplicationService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateProductTemplateAttributeValueCommand      = composable.CreateCommand
	UpdateProductTemplateAttributeValueCommand      = composable.UpdateCommand
	DeleteProductTemplateAttributeValueCommand      = composable.DeleteCommand
	SetProductTemplateAttributeValueArchivedCommand = composable.SetArchivedCommand
	GetProductTemplateAttributeValueByIdQuery       = composable.GetByIdQuery
	SearchProductTemplateAttributeValuesQuery       = composable.SearchQuery
	ProductTemplateAttributeValueExistsQuery        = composable.ExistsQuery
)

type (
	CreateProductTemplateAttributeValueResult      = composable.CreateResult
	UpdateProductTemplateAttributeValueResult      = composable.MutateResult
	DeleteProductTemplateAttributeValueResult      = composable.MutateResult
	SetProductTemplateAttributeValueArchivedResult = composable.MutateResult
	GetProductTemplateAttributeValueByIdResult     = composable.GetOneResult
	SearchProductTemplateAttributeValuesResult     = composable.SearchResult
	ProductTemplateAttributeValueExistsResult      = composable.ExistsResult
)

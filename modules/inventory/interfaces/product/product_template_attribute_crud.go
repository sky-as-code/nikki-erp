package product

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// ProductTemplateAttributeRepository reads and writes product_template_attribute rows.
type ProductTemplateAttributeRepository interface {
	composable.CrudRepository
}

// ProductTemplateAttributeDomainService is the CRUD of the resource plus its own rules.
type ProductTemplateAttributeDomainService interface {
	composable.CrudDomainService
}

// ProductTemplateAttributeApplicationService is the authorized surface the REST handler serves.
type ProductTemplateAttributeApplicationService interface {
	composable.CrudApplicationService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateProductTemplateAttributeCommand      = composable.CreateCommand
	UpdateProductTemplateAttributeCommand      = composable.UpdateCommand
	DeleteProductTemplateAttributeCommand      = composable.DeleteCommand
	SetProductTemplateAttributeArchivedCommand = composable.SetArchivedCommand
	GetProductTemplateAttributeByIdQuery       = composable.GetByIdQuery
	SearchProductTemplateAttributesQuery       = composable.SearchQuery
	ProductTemplateAttributeExistsQuery        = composable.ExistsQuery
)

type (
	CreateProductTemplateAttributeResult      = composable.CreateResult
	UpdateProductTemplateAttributeResult      = composable.MutateResult
	DeleteProductTemplateAttributeResult      = composable.MutateResult
	SetProductTemplateAttributeArchivedResult = composable.MutateResult
	GetProductTemplateAttributeByIdResult     = composable.GetOneResult
	SearchProductTemplateAttributesResult     = composable.SearchResult
	ProductTemplateAttributeExistsResult      = composable.ExistsResult
)

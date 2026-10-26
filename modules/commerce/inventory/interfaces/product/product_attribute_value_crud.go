package product

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// ProductAttributeValueRepository reads and writes product_attribute_value rows.
type ProductAttributeValueRepository interface {
	composable.CrudRepository
}

// ProductAttributeValueDomainService is the CRUD of the resource plus its own rules.
type ProductAttributeValueDomainService interface {
	composable.CrudDomainService
}

// ProductAttributeValueApplicationService is the authorized surface the REST handler serves.
type ProductAttributeValueApplicationService interface {
	composable.CrudApplicationService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateProductAttributeValueCommand      = composable.CreateCommand
	UpdateProductAttributeValueCommand      = composable.UpdateCommand
	DeleteProductAttributeValueCommand      = composable.DeleteCommand
	SetProductAttributeValueArchivedCommand = composable.SetArchivedCommand
	GetProductAttributeValueByIdQuery       = composable.GetByIdQuery
	SearchProductAttributeValuesQuery       = composable.SearchQuery
	ProductAttributeValueExistsQuery        = composable.ExistsQuery
)

type (
	CreateProductAttributeValueResult      = composable.CreateResult
	UpdateProductAttributeValueResult      = composable.MutateResult
	DeleteProductAttributeValueResult      = composable.MutateResult
	SetProductAttributeValueArchivedResult = composable.MutateResult
	GetProductAttributeValueByIdResult     = composable.GetOneResult
	SearchProductAttributeValuesResult     = composable.SearchResult
	ProductAttributeValueExistsResult      = composable.ExistsResult
)

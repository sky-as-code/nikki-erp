package product

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// ProductAttributeValueRepository reads and writes product_attribute_value rows.
type ProductAttributeValueRepository interface {
	dynamicresource.CrudRepository
}

// ProductAttributeValueDomainService is the CRUD of the resource plus its own rules.
type ProductAttributeValueDomainService interface {
	dynamicresource.CrudDomainService
}

// ProductAttributeValueApplicationService is the authorized surface the REST handler serves.
type ProductAttributeValueApplicationService interface {
	dynamicresource.CrudApplicationService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateProductAttributeValueCommand      = dynamicresource.CreateCommand
	UpdateProductAttributeValueCommand      = dynamicresource.UpdateCommand
	DeleteProductAttributeValueCommand      = dynamicresource.DeleteCommand
	SetProductAttributeValueArchivedCommand = dynamicresource.SetArchivedCommand
	GetProductAttributeValueByIdQuery       = dynamicresource.GetByIdQuery
	SearchProductAttributeValuesQuery       = dynamicresource.SearchQuery
	ProductAttributeValueExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateProductAttributeValueResult      = dynamicresource.CreateResult
	UpdateProductAttributeValueResult      = dynamicresource.MutateResult
	DeleteProductAttributeValueResult      = dynamicresource.MutateResult
	SetProductAttributeValueArchivedResult = dynamicresource.MutateResult
	GetProductAttributeValueByIdResult     = dynamicresource.GetOneResult
	SearchProductAttributeValuesResult     = dynamicresource.SearchResult
	ProductAttributeValueExistsResult      = dynamicresource.ExistsResult
)

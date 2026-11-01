package product

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// ProductVariantAttributeValueRepository reads and writes product_variant_attribute_value rows.
type ProductVariantAttributeValueRepository interface {
	dynamicresource.CrudRepository
}

// ProductVariantAttributeValueDomainService is the CRUD of the resource plus its own rules.
type ProductVariantAttributeValueDomainService interface {
	dynamicresource.CrudDomainService
}

// ProductVariantAttributeValueApplicationService is the authorized surface the REST handler serves.
type ProductVariantAttributeValueApplicationService interface {
	dynamicresource.CrudApplicationService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateProductVariantAttributeValueCommand      = dynamicresource.CreateCommand
	UpdateProductVariantAttributeValueCommand      = dynamicresource.UpdateCommand
	DeleteProductVariantAttributeValueCommand      = dynamicresource.DeleteCommand
	SetProductVariantAttributeValueArchivedCommand = dynamicresource.SetArchivedCommand
	GetProductVariantAttributeValueByIdQuery       = dynamicresource.GetByIdQuery
	SearchProductVariantAttributeValuesQuery       = dynamicresource.SearchQuery
	ProductVariantAttributeValueExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateProductVariantAttributeValueResult      = dynamicresource.CreateResult
	UpdateProductVariantAttributeValueResult      = dynamicresource.MutateResult
	DeleteProductVariantAttributeValueResult      = dynamicresource.MutateResult
	SetProductVariantAttributeValueArchivedResult = dynamicresource.MutateResult
	GetProductVariantAttributeValueByIdResult     = dynamicresource.GetOneResult
	SearchProductVariantAttributeValuesResult     = dynamicresource.SearchResult
	ProductVariantAttributeValueExistsResult      = dynamicresource.ExistsResult
)

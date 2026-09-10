package product

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// ProductVariantAttributeValueRepository reads and writes product_variant_attribute_value rows.
type ProductVariantAttributeValueRepository interface {
	composable.CrudRepository
}

// ProductVariantAttributeValueDomainService is the CRUD of the resource plus its own rules.
type ProductVariantAttributeValueDomainService interface {
	composable.CrudDomainService
}

// ProductVariantAttributeValueApplicationService is the authorized surface the REST handler serves.
type ProductVariantAttributeValueApplicationService interface {
	composable.CrudApplicationService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateProductVariantAttributeValueCommand      = composable.CreateCommand
	UpdateProductVariantAttributeValueCommand      = composable.UpdateCommand
	DeleteProductVariantAttributeValueCommand      = composable.DeleteCommand
	SetProductVariantAttributeValueArchivedCommand = composable.SetArchivedCommand
	GetProductVariantAttributeValueByIdQuery       = composable.GetByIdQuery
	SearchProductVariantAttributeValuesQuery       = composable.SearchQuery
	ProductVariantAttributeValueExistsQuery        = composable.ExistsQuery
)

type (
	CreateProductVariantAttributeValueResult      = composable.CreateResult
	UpdateProductVariantAttributeValueResult      = composable.MutateResult
	DeleteProductVariantAttributeValueResult      = composable.MutateResult
	SetProductVariantAttributeValueArchivedResult = composable.MutateResult
	GetProductVariantAttributeValueByIdResult     = composable.GetOneResult
	SearchProductVariantAttributeValuesResult     = composable.SearchResult
	ProductVariantAttributeValueExistsResult      = composable.ExistsResult
)

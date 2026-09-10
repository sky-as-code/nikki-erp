package product

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// ProductAttributeRepository reads and writes product_attribute rows.
type ProductAttributeRepository interface {
	composable.CrudRepository
}

// ProductAttributeDomainService is the CRUD of the resource plus its own rules.
type ProductAttributeDomainService interface {
	composable.CrudDomainService
}

// ProductAttributeApplicationService is the authorized surface the REST handler serves.
type ProductAttributeApplicationService interface {
	composable.CrudApplicationService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateProductAttributeCommand      = composable.CreateCommand
	UpdateProductAttributeCommand      = composable.UpdateCommand
	DeleteProductAttributeCommand      = composable.DeleteCommand
	SetProductAttributeArchivedCommand = composable.SetArchivedCommand
	GetProductAttributeByIdQuery       = composable.GetByIdQuery
	SearchProductAttributesQuery       = composable.SearchQuery
	ProductAttributeExistsQuery        = composable.ExistsQuery
)

type (
	CreateProductAttributeResult      = composable.CreateResult
	UpdateProductAttributeResult      = composable.MutateResult
	DeleteProductAttributeResult      = composable.MutateResult
	SetProductAttributeArchivedResult = composable.MutateResult
	GetProductAttributeByIdResult     = composable.GetOneResult
	SearchProductAttributesResult     = composable.SearchResult
	ProductAttributeExistsResult      = composable.ExistsResult
)

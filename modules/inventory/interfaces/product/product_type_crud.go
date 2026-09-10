package product

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// ProductTypeRepository reads and writes product_type rows.
type ProductTypeRepository interface {
	composable.CrudRepository
}

// ProductTypeDomainService is the CRUD of the resource plus its own rules.
type ProductTypeDomainService interface {
	composable.CrudDomainService
}

// ProductTypeApplicationService is the authorized surface the REST handler serves.
type ProductTypeApplicationService interface {
	composable.CrudApplicationService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateProductTypeCommand      = composable.CreateCommand
	UpdateProductTypeCommand      = composable.UpdateCommand
	DeleteProductTypeCommand      = composable.DeleteCommand
	SetProductTypeArchivedCommand = composable.SetArchivedCommand
	GetProductTypeByIdQuery       = composable.GetByIdQuery
	SearchProductTypesQuery       = composable.SearchQuery
	ProductTypeExistsQuery        = composable.ExistsQuery
)

type (
	CreateProductTypeResult      = composable.CreateResult
	UpdateProductTypeResult      = composable.MutateResult
	DeleteProductTypeResult      = composable.MutateResult
	SetProductTypeArchivedResult = composable.MutateResult
	GetProductTypeByIdResult     = composable.GetOneResult
	SearchProductTypesResult     = composable.SearchResult
	ProductTypeExistsResult      = composable.ExistsResult
)

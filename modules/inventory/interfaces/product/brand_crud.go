package product

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// BrandRepository reads and writes brand rows.
type BrandRepository interface {
	composable.CrudRepository
}

// BrandDomainService is the CRUD of the resource plus its own rules.
type BrandDomainService interface {
	composable.CrudDomainService
}

// BrandApplicationService is the authorized surface the REST handler serves.
type BrandApplicationService interface {
	composable.CrudApplicationService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateBrandCommand      = composable.CreateCommand
	UpdateBrandCommand      = composable.UpdateCommand
	DeleteBrandCommand      = composable.DeleteCommand
	SetBrandArchivedCommand = composable.SetArchivedCommand
	GetBrandByIdQuery       = composable.GetByIdQuery
	SearchBrandsQuery       = composable.SearchQuery
	BrandExistsQuery        = composable.ExistsQuery
)

type (
	CreateBrandResult      = composable.CreateResult
	UpdateBrandResult      = composable.MutateResult
	DeleteBrandResult      = composable.MutateResult
	SetBrandArchivedResult = composable.MutateResult
	GetBrandByIdResult     = composable.GetOneResult
	SearchBrandsResult     = composable.SearchResult
	BrandExistsResult      = composable.ExistsResult
)

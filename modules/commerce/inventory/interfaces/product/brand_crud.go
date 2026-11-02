package product

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// BrandRepository reads and writes brand rows.
type BrandRepository interface {
	dynamicresource.CrudRepository
}

// BrandDomainService is the CRUD of the resource plus its own rules.
type BrandDomainService interface {
	dynamicresource.CrudDomainService
}

// BrandApplicationService is the authorized surface the REST handler serves.
type BrandApplicationService interface {
	dynamicresource.CrudApplicationService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateBrandCommand      = dynamicresource.CreateCommand
	UpdateBrandCommand      = dynamicresource.UpdateCommand
	DeleteBrandCommand      = dynamicresource.DeleteCommand
	SetBrandArchivedCommand = dynamicresource.SetArchivedCommand
	GetBrandByIdQuery       = dynamicresource.GetByIdQuery
	SearchBrandsQuery       = dynamicresource.SearchQuery
	BrandExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateBrandResult      = dynamicresource.CreateResult
	UpdateBrandResult      = dynamicresource.MutateResult
	DeleteBrandResult      = dynamicresource.MutateResult
	SetBrandArchivedResult = dynamicresource.MutateResult
	GetBrandByIdResult     = dynamicresource.GetOneResult
	SearchBrandsResult     = dynamicresource.SearchResult
	BrandExistsResult      = dynamicresource.ExistsResult
)

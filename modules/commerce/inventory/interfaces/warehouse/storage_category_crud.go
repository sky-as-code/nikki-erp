package warehouse

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// StorageCategoryRepository reads and writes storage_category rows.
type StorageCategoryRepository interface {
	composable.CrudRepository
}

// StorageCategoryDomainService is the CRUD of the resource plus its own rules.
type StorageCategoryDomainService interface {
	composable.CrudDomainService
}

// StorageCategoryApplicationService is the authorized surface the REST handler serves.
type StorageCategoryApplicationService interface {
	composable.CrudApplicationService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateStorageCategoryCommand      = composable.CreateCommand
	UpdateStorageCategoryCommand      = composable.UpdateCommand
	DeleteStorageCategoryCommand      = composable.DeleteCommand
	SetStorageCategoryArchivedCommand = composable.SetArchivedCommand
	GetStorageCategoryByIdQuery       = composable.GetByIdQuery
	SearchStorageCategoriesQuery      = composable.SearchQuery
	StorageCategoryExistsQuery        = composable.ExistsQuery
)

type (
	CreateStorageCategoryResult      = composable.CreateResult
	UpdateStorageCategoryResult      = composable.MutateResult
	DeleteStorageCategoryResult      = composable.MutateResult
	SetStorageCategoryArchivedResult = composable.MutateResult
	GetStorageCategoryByIdResult     = composable.GetOneResult
	SearchStorageCategoriesResult    = composable.SearchResult
	StorageCategoryExistsResult      = composable.ExistsResult
)

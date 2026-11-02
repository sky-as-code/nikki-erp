package warehouse

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// StorageCategoryRepository reads and writes storage_category rows.
type StorageCategoryRepository interface {
	dynamicresource.CrudRepository
}

// StorageCategoryDomainService is the CRUD of the resource plus its own rules.
type StorageCategoryDomainService interface {
	dynamicresource.CrudDomainService
}

// StorageCategoryApplicationService is the authorized surface the REST handler serves.
type StorageCategoryApplicationService interface {
	dynamicresource.CrudApplicationService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateStorageCategoryCommand      = dynamicresource.CreateCommand
	UpdateStorageCategoryCommand      = dynamicresource.UpdateCommand
	DeleteStorageCategoryCommand      = dynamicresource.DeleteCommand
	SetStorageCategoryArchivedCommand = dynamicresource.SetArchivedCommand
	GetStorageCategoryByIdQuery       = dynamicresource.GetByIdQuery
	SearchStorageCategoriesQuery      = dynamicresource.SearchQuery
	StorageCategoryExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateStorageCategoryResult      = dynamicresource.CreateResult
	UpdateStorageCategoryResult      = dynamicresource.MutateResult
	DeleteStorageCategoryResult      = dynamicresource.MutateResult
	SetStorageCategoryArchivedResult = dynamicresource.MutateResult
	GetStorageCategoryByIdResult     = dynamicresource.GetOneResult
	SearchStorageCategoriesResult    = dynamicresource.SearchResult
	StorageCategoryExistsResult      = dynamicresource.ExistsResult
)

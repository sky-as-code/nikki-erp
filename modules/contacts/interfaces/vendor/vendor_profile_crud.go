package vendor

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// VendorProfileRepository reads and writes vendor profile rows.
type VendorProfileRepository interface {
	composable.CrudRepository
}

// VendorProfileApplicationService is the whole application layer of the vendor profile resource:
// the authorized CRUD the REST handler serves, plus the cross-module capability.
type VendorProfileApplicationService interface {
	composable.CrudApplicationService
	VendorAppService
}

type (
	CreateVendorProfileCommand      = composable.CreateCommand
	UpdateVendorProfileCommand      = composable.UpdateCommand
	DeleteVendorProfileCommand      = composable.DeleteCommand
	SetVendorProfileArchivedCommand = composable.SetArchivedCommand
	GetVendorProfileByIdQuery       = composable.GetByIdQuery
	SearchVendorProfilesQuery       = composable.SearchQuery
	VendorProfileExistsQuery        = composable.ExistsQuery
)

type (
	CreateVendorProfileResult      = composable.CreateResult
	UpdateVendorProfileResult      = composable.MutateResult
	DeleteVendorProfileResult      = composable.MutateResult
	SetVendorProfileArchivedResult = composable.MutateResult
	GetVendorProfileByIdResult     = composable.GetOneResult
	SearchVendorProfilesResult     = composable.SearchResult
	VendorProfileExistsResult      = composable.ExistsResult
)

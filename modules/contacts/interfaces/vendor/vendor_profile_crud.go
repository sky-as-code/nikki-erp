package vendor

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// VendorProfileRepository reads and writes vendor profile rows.
type VendorProfileRepository interface {
	dynamicresource.CrudRepository
}

// VendorProfileApplicationService is the whole application layer of the vendor profile resource:
// the authorized CRUD the REST handler serves, plus the cross-module capability.
type VendorProfileApplicationService interface {
	dynamicresource.CrudApplicationService
	VendorAppService
}

type (
	CreateVendorProfileCommand      = dynamicresource.CreateCommand
	UpdateVendorProfileCommand      = dynamicresource.UpdateCommand
	DeleteVendorProfileCommand      = dynamicresource.DeleteCommand
	SetVendorProfileArchivedCommand = dynamicresource.SetArchivedCommand
	GetVendorProfileByIdQuery       = dynamicresource.GetByIdQuery
	SearchVendorProfilesQuery       = dynamicresource.SearchQuery
	VendorProfileExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateVendorProfileResult      = dynamicresource.CreateResult
	UpdateVendorProfileResult      = dynamicresource.MutateResult
	DeleteVendorProfileResult      = dynamicresource.MutateResult
	SetVendorProfileArchivedResult = dynamicresource.MutateResult
	GetVendorProfileByIdResult     = dynamicresource.GetOneResult
	SearchVendorProfilesResult     = dynamicresource.SearchResult
	VendorProfileExistsResult      = dynamicresource.ExistsResult
)

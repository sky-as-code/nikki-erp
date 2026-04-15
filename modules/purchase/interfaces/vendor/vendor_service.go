package vendor

import corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

type VendorService interface {
	CreateVendor(ctx corectx.Context, cmd CreateVendorCommand) (*CreateVendorResult, error)
	DeleteVendor(ctx corectx.Context, cmd DeleteVendorCommand) (*DeleteVendorResult, error)
	VendorExists(ctx corectx.Context, query VendorExistsQuery) (*VendorExistsResult, error)
	GetVendor(ctx corectx.Context, query GetVendorQuery) (*GetVendorResult, error)
	SearchVendors(ctx corectx.Context, query SearchVendorsQuery) (*SearchVendorsResult, error)
	SetVendorIsArchived(ctx corectx.Context, cmd SetVendorIsArchivedCommand) (*SetVendorIsArchivedResult, error)
	UpdateVendor(ctx corectx.Context, cmd UpdateVendorCommand) (*UpdateVendorResult, error)
}

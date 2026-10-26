// Package vendorprice declares the CRUD layers of the vendor price list.
package vendorprice

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// VendorProductPriceCrudRepository reads and writes vendor product prices.
type VendorProductPriceCrudRepository interface {
	composable.CrudRepository
}

// VendorProductPriceCrudDomainService is the CRUD of the resource. Create and update validate the record as it will be;
// unarchiving revalidates it.
type VendorProductPriceCrudDomainService interface {
	composable.CrudDomainService
}

// VendorProductPriceCrudApplicationService is the authorized surface the REST handler serves.
type VendorProductPriceCrudApplicationService interface {
	composable.CrudApplicationService
}

type (
	CreateVendorProductPriceCommand      = composable.CreateCommand
	UpdateVendorProductPriceCommand      = composable.UpdateCommand
	DeleteVendorProductPriceCommand      = composable.DeleteCommand
	SetVendorProductPriceArchivedCommand = composable.SetArchivedCommand
	GetVendorProductPriceByIdQuery       = composable.GetByIdQuery
	SearchVendorProductPricesQuery       = composable.SearchQuery
	VendorProductPriceExistsQuery        = composable.ExistsQuery
)

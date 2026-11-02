// Package vendorprice declares the CRUD layers of the vendor price list.
package vendorprice

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// VendorProductPriceCrudRepository reads and writes vendor product prices.
type VendorProductPriceCrudRepository interface {
	dynamicresource.CrudRepository
}

// VendorProductPriceCrudDomainService is the CRUD of the resource. Create and update validate the record as it will be;
// unarchiving revalidates it.
type VendorProductPriceCrudDomainService interface {
	dynamicresource.CrudDomainService
}

// VendorProductPriceCrudApplicationService is the authorized surface the REST handler serves.
type VendorProductPriceCrudApplicationService interface {
	dynamicresource.CrudApplicationService
}

type (
	CreateVendorProductPriceCommand      = dynamicresource.CreateCommand
	UpdateVendorProductPriceCommand      = dynamicresource.UpdateCommand
	DeleteVendorProductPriceCommand      = dynamicresource.DeleteCommand
	SetVendorProductPriceArchivedCommand = dynamicresource.SetArchivedCommand
	GetVendorProductPriceByIdQuery       = dynamicresource.GetByIdQuery
	SearchVendorProductPricesQuery       = dynamicresource.SearchQuery
	VendorProductPriceExistsQuery        = dynamicresource.ExistsQuery
)

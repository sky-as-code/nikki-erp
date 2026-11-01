package repository

import (
	itVendorPrice "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/vendorprice"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

func NewVendorProductPriceRepository(base dynamicresource.CrudRepository) itVendorPrice.VendorProductPriceCrudRepository {
	return &VendorProductPriceRepositoryImpl{CrudRepository: base}
}

type VendorProductPriceRepositoryImpl struct {
	dynamicresource.CrudRepository
}

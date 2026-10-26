package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itVendorPrice "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/vendorprice"
)

func NewVendorProductPriceRepository(base composable.CrudRepository) itVendorPrice.VendorProductPriceCrudRepository {
	return &VendorProductPriceRepositoryImpl{CrudRepository: base}
}

type VendorProductPriceRepositoryImpl struct {
	composable.CrudRepository
}

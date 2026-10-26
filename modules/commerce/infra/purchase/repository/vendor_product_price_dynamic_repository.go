package repository

import (
	itVendorPrice "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/vendorprice"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

func NewVendorProductPriceRepository(base composable.CrudRepository) itVendorPrice.VendorProductPriceCrudRepository {
	return &VendorProductPriceRepositoryImpl{CrudRepository: base}
}

type VendorProductPriceRepositoryImpl struct {
	composable.CrudRepository
}

package repository

import (
	itCatalog "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/catalog"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

func NewSalesPricelistRepository(base dynamicresource.CrudRepository) itCatalog.SalesPricelistRepository {
	return &SalesPricelistRepositoryImpl{CrudRepository: base}
}

type SalesPricelistRepositoryImpl struct {
	dynamicresource.CrudRepository
}

func NewSalesPricelistItemRepository(base dynamicresource.CrudRepository) itCatalog.SalesPricelistItemRepository {
	return &SalesPricelistItemRepositoryImpl{CrudRepository: base}
}

type SalesPricelistItemRepositoryImpl struct {
	dynamicresource.CrudRepository
}

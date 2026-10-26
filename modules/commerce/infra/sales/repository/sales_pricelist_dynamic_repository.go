package repository

import (
	itCatalog "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/catalog"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

func NewSalesPricelistRepository(base composable.CrudRepository) itCatalog.SalesPricelistRepository {
	return &SalesPricelistRepositoryImpl{CrudRepository: base}
}

type SalesPricelistRepositoryImpl struct {
	composable.CrudRepository
}

func NewSalesPricelistItemRepository(base composable.CrudRepository) itCatalog.SalesPricelistItemRepository {
	return &SalesPricelistItemRepositoryImpl{CrudRepository: base}
}

type SalesPricelistItemRepositoryImpl struct {
	composable.CrudRepository
}

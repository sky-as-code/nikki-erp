package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itCatalog "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/catalog"
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

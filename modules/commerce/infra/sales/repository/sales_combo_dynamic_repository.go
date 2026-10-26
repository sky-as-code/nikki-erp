package repository

import (
	itCatalog "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/catalog"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

func NewSalesComboRepository(base composable.CrudRepository) itCatalog.SalesComboRepository {
	return &SalesComboRepositoryImpl{CrudRepository: base}
}

type SalesComboRepositoryImpl struct {
	composable.CrudRepository
}

func NewSalesComboComponentRepository(base composable.CrudRepository) itCatalog.SalesComboComponentRepository {
	return &SalesComboComponentRepositoryImpl{CrudRepository: base}
}

type SalesComboComponentRepositoryImpl struct {
	composable.CrudRepository
}

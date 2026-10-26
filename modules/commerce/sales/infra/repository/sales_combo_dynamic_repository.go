package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itCatalog "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/catalog"
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

package repository

import (
	itCatalog "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/catalog"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

func NewSalesComboRepository(base dynamicresource.CrudRepository) itCatalog.SalesComboRepository {
	return &SalesComboRepositoryImpl{CrudRepository: base}
}

type SalesComboRepositoryImpl struct {
	dynamicresource.CrudRepository
}

func NewSalesComboComponentRepository(base dynamicresource.CrudRepository) itCatalog.SalesComboComponentRepository {
	return &SalesComboComponentRepositoryImpl{CrudRepository: base}
}

type SalesComboComponentRepositoryImpl struct {
	dynamicresource.CrudRepository
}

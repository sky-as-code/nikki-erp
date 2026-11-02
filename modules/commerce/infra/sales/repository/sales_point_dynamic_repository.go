package repository

import (
	itCatalog "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/catalog"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

func NewSalesPointRepository(base dynamicresource.CrudRepository) itCatalog.SalesPointRepository {
	return &SalesPointRepositoryImpl{CrudRepository: base}
}

type SalesPointRepositoryImpl struct {
	dynamicresource.CrudRepository
}

package repository

import (
	itCatalog "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/catalog"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

func NewSalesPointRepository(base composable.CrudRepository) itCatalog.SalesPointRepository {
	return &SalesPointRepositoryImpl{CrudRepository: base}
}

type SalesPointRepositoryImpl struct {
	composable.CrudRepository
}

package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itCatalog "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/catalog"
)

func NewSalesPointRepository(base composable.CrudRepository) itCatalog.SalesPointRepository {
	return &SalesPointRepositoryImpl{CrudRepository: base}
}

type SalesPointRepositoryImpl struct {
	composable.CrudRepository
}

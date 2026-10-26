package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itCatalog "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/catalog"
)

func NewSalesChannelRepository(base composable.CrudRepository) itCatalog.SalesChannelRepository {
	return &SalesChannelRepositoryImpl{CrudRepository: base}
}

type SalesChannelRepositoryImpl struct {
	composable.CrudRepository
}

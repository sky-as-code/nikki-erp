package repository

import (
	itCatalog "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/catalog"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

func NewSalesChannelRepository(base dynamicresource.CrudRepository) itCatalog.SalesChannelRepository {
	return &SalesChannelRepositoryImpl{CrudRepository: base}
}

type SalesChannelRepositoryImpl struct {
	dynamicresource.CrudRepository
}

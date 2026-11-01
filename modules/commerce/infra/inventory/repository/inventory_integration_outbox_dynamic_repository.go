package repository

import (
	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

func NewInventoryIntegrationOutboxRepository(
	base dynamicresource.CrudRepository,
) itStock.InventoryIntegrationOutboxRepository {
	return &InventoryIntegrationOutboxRepositoryImpl{CrudRepository: base}
}

type InventoryIntegrationOutboxRepositoryImpl struct {
	dynamicresource.CrudRepository
}

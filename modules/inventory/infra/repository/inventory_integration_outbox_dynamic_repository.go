package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

func NewInventoryIntegrationOutboxRepository(base composable.CrudRepository) itStock.InventoryIntegrationOutboxRepository {
	return &InventoryIntegrationOutboxRepositoryImpl{CrudRepository: base}
}

type InventoryIntegrationOutboxRepositoryImpl struct {
	composable.CrudRepository
}

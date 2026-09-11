package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itIntegration "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/integration"
)

func NewSalesManualDiscountRepository(base composable.CrudRepository) itIntegration.SalesManualDiscountRepository {
	return &SalesManualDiscountRepositoryImpl{CrudRepository: base}
}

type SalesManualDiscountRepositoryImpl struct {
	composable.CrudRepository
}

func NewSalesIntegrationOutboxRepository(base composable.CrudRepository) itIntegration.SalesIntegrationOutboxRepository {
	return &SalesIntegrationOutboxRepositoryImpl{CrudRepository: base}
}

type SalesIntegrationOutboxRepositoryImpl struct {
	composable.CrudRepository
}

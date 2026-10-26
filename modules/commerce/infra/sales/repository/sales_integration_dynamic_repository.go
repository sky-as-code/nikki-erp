package repository

import (
	itIntegration "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/integration"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
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

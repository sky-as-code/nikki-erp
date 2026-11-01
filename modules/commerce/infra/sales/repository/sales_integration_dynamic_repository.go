package repository

import (
	itIntegration "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/integration"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

func NewSalesManualDiscountRepository(base dynamicresource.CrudRepository) itIntegration.SalesManualDiscountRepository {
	return &SalesManualDiscountRepositoryImpl{CrudRepository: base}
}

type SalesManualDiscountRepositoryImpl struct {
	dynamicresource.CrudRepository
}

func NewSalesIntegrationOutboxRepository(
	base dynamicresource.CrudRepository,
) itIntegration.SalesIntegrationOutboxRepository {
	return &SalesIntegrationOutboxRepositoryImpl{CrudRepository: base}
}

type SalesIntegrationOutboxRepositoryImpl struct {
	dynamicresource.CrudRepository
}

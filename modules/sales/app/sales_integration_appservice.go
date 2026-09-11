package app

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itIntegration "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/integration"
)

func NewSalesManualDiscountApplicationService(base composable.CrudApplicationService) itIntegration.SalesManualDiscountApplicationService {
	return &SalesManualDiscountApplicationServiceImpl{CrudApplicationService: base}
}

type SalesManualDiscountApplicationServiceImpl struct {
	composable.CrudApplicationService
}

func NewSalesIntegrationOutboxApplicationService(base composable.CrudApplicationService) itIntegration.SalesIntegrationOutboxApplicationService {
	return &SalesIntegrationOutboxApplicationServiceImpl{CrudApplicationService: base}
}

type SalesIntegrationOutboxApplicationServiceImpl struct {
	composable.CrudApplicationService
}

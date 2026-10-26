package app

import (
	itIntegration "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/integration"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
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

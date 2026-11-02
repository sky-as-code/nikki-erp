package app

import (
	itIntegration "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/integration"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

func NewSalesManualDiscountApplicationService(base dynamicresource.CrudApplicationService) itIntegration.SalesManualDiscountApplicationService {
	return &SalesManualDiscountApplicationServiceImpl{CrudApplicationService: base}
}

type SalesManualDiscountApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
}

func NewSalesIntegrationOutboxApplicationService(base dynamicresource.CrudApplicationService) itIntegration.SalesIntegrationOutboxApplicationService {
	return &SalesIntegrationOutboxApplicationServiceImpl{CrudApplicationService: base}
}

type SalesIntegrationOutboxApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
}

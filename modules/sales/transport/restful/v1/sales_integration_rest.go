package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

const (
	SalesManualDiscountEngineName    = "dynengine_sales_manual_discount"
	SalesIntegrationOutboxEngineName = "dynengine_sales_integration_outbox"
)

type salesManualDiscountRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_manual_discount"`
}

func NewSalesManualDiscountRest(params salesManualDiscountRestParams) *SalesManualDiscountRest {
	rest := &SalesManualDiscountRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesManualDiscountRest struct {
	composable.CrudRestBase
}

type salesIntegrationOutboxRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_integration_outbox"`
}

func NewSalesIntegrationOutboxRest(params salesIntegrationOutboxRestParams) *SalesIntegrationOutboxRest {
	rest := &SalesIntegrationOutboxRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesIntegrationOutboxRest struct {
	composable.CrudRestBase
}

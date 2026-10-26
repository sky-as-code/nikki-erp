package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itFulfillment "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/fulfillment"
)

const (
	SalesOrderFulfillmentEngineName = "dynengine_sales_order_fulfillment"
	SalesOrderFulfillmentItemEngineName = "dynengine_sales_order_fulfillment_item"
	SalesFulfillmentAttemptEngineName = "dynengine_sales_fulfillment_attempt"
	SalesFulfillmentAttemptItemEngineName = "dynengine_sales_fulfillment_attempt_item"
	SalesFulfillmentTargetChangeEngineName = "dynengine_sales_fulfillment_target_change"
)

type salesOrderFulfillmentRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_order_fulfillment"`
}

func NewSalesOrderFulfillmentRest(params salesOrderFulfillmentRestParams) *SalesOrderFulfillmentRest {
	rest := &SalesOrderFulfillmentRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.fulfillmentSvc = params.Engine.ApplicationService().(itFulfillment.SalesOrderFulfillmentApplicationService)
	return rest
}

type SalesOrderFulfillmentRest struct {
	composable.CrudRestBase
	fulfillmentSvc itFulfillment.SalesOrderFulfillmentApplicationService
}

type salesOrderFulfillmentItemRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_order_fulfillment_item"`
}

func NewSalesOrderFulfillmentItemRest(params salesOrderFulfillmentItemRestParams) *SalesOrderFulfillmentItemRest {
	rest := &SalesOrderFulfillmentItemRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesOrderFulfillmentItemRest struct {
	composable.CrudRestBase
}

type salesFulfillmentAttemptRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_fulfillment_attempt"`
}

func NewSalesFulfillmentAttemptRest(params salesFulfillmentAttemptRestParams) *SalesFulfillmentAttemptRest {
	rest := &SalesFulfillmentAttemptRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesFulfillmentAttemptRest struct {
	composable.CrudRestBase
}

type salesFulfillmentAttemptItemRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_fulfillment_attempt_item"`
}

func NewSalesFulfillmentAttemptItemRest(params salesFulfillmentAttemptItemRestParams) *SalesFulfillmentAttemptItemRest {
	rest := &SalesFulfillmentAttemptItemRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesFulfillmentAttemptItemRest struct {
	composable.CrudRestBase
}

type salesFulfillmentTargetChangeRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_fulfillment_target_change"`
}

func NewSalesFulfillmentTargetChangeRest(params salesFulfillmentTargetChangeRestParams) *SalesFulfillmentTargetChangeRest {
	rest := &SalesFulfillmentTargetChangeRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesFulfillmentTargetChangeRest struct {
	composable.CrudRestBase
}

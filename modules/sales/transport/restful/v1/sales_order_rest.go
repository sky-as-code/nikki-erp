package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itOrder "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/order"
)

const (
	SalesOrderEngineName              = "dynengine_sales_order"
	SalesOrderLineEngineName          = "dynengine_sales_order_line"
	SalesOrderLineComponentEngineName = "dynengine_sales_order_line_component"
	SalesOrderAdjustmentEngineName    = "dynengine_sales_order_adjustment"
	SalesOrderEventEngineName         = "dynengine_sales_order_event"
)

type salesOrderRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_order"`
}

func NewSalesOrderRest(params salesOrderRestParams) *SalesOrderRest {
	rest := &SalesOrderRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.orderSvc = params.Engine.ApplicationService().(itOrder.SalesOrderApplicationService)
	return rest
}

type SalesOrderRest struct {
	composable.CrudRestBase
	orderSvc itOrder.SalesOrderApplicationService
}

type salesOrderLineRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_order_line"`
}

func NewSalesOrderLineRest(params salesOrderLineRestParams) *SalesOrderLineRest {
	rest := &SalesOrderLineRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesOrderLineRest struct {
	composable.CrudRestBase
}

type salesOrderLineComponentRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_order_line_component"`
}

func NewSalesOrderLineComponentRest(params salesOrderLineComponentRestParams) *SalesOrderLineComponentRest {
	rest := &SalesOrderLineComponentRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesOrderLineComponentRest struct {
	composable.CrudRestBase
}

type salesOrderAdjustmentRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_order_adjustment"`
}

func NewSalesOrderAdjustmentRest(params salesOrderAdjustmentRestParams) *SalesOrderAdjustmentRest {
	rest := &SalesOrderAdjustmentRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesOrderAdjustmentRest struct {
	composable.CrudRestBase
}

type salesOrderEventRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_order_event"`
}

func NewSalesOrderEventRest(params salesOrderEventRestParams) *SalesOrderEventRest {
	rest := &SalesOrderEventRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesOrderEventRest struct {
	composable.CrudRestBase
}

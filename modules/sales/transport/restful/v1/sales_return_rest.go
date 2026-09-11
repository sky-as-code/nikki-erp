package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itReturns "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/returns"
)

const (
	SalesReturnEngineName = "dynengine_sales_return"
	SalesReturnLineEngineName = "dynengine_sales_return_line"
	SalesRefundPaymentEngineName = "dynengine_sales_refund_payment"
)

type salesReturnRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_return"`
}

func NewSalesReturnRest(params salesReturnRestParams) *SalesReturnRest {
	rest := &SalesReturnRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.returnSvc = params.Engine.ApplicationService().(itReturns.SalesReturnApplicationService)
	return rest
}

type SalesReturnRest struct {
	composable.CrudRestBase
	returnSvc itReturns.SalesReturnApplicationService
}

type salesReturnLineRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_return_line"`
}

func NewSalesReturnLineRest(params salesReturnLineRestParams) *SalesReturnLineRest {
	rest := &SalesReturnLineRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesReturnLineRest struct {
	composable.CrudRestBase
}

type salesRefundPaymentRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_refund_payment"`
}

func NewSalesRefundPaymentRest(params salesRefundPaymentRestParams) *SalesRefundPaymentRest {
	rest := &SalesRefundPaymentRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesRefundPaymentRest struct {
	composable.CrudRestBase
}

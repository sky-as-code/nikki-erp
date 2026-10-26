package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itBilling "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/billing"
)

const (
	SalesBillEngineName = "dynengine_sales_bill"
	SalesBillLineEngineName = "dynengine_sales_bill_line"
	SalesBillRelationEngineName = "dynengine_sales_bill_relation"
	SalesPaymentEngineName = "dynengine_sales_payment"
	SalesFulfillmentRequestEngineName = "dynengine_sales_fulfillment_request"
	SalesFulfillmentRequestLineEngineName = "dynengine_sales_fulfillment_request_line"
)

type salesBillRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_bill"`
}

func NewSalesBillRest(params salesBillRestParams) *SalesBillRest {
	rest := &SalesBillRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.billSvc = params.Engine.ApplicationService().(itBilling.SalesBillApplicationService)
	return rest
}

type SalesBillRest struct {
	composable.CrudRestBase
	billSvc itBilling.SalesBillApplicationService
}

type salesBillLineRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_bill_line"`
}

func NewSalesBillLineRest(params salesBillLineRestParams) *SalesBillLineRest {
	rest := &SalesBillLineRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesBillLineRest struct {
	composable.CrudRestBase
}

type salesBillRelationRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_bill_relation"`
}

func NewSalesBillRelationRest(params salesBillRelationRestParams) *SalesBillRelationRest {
	rest := &SalesBillRelationRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesBillRelationRest struct {
	composable.CrudRestBase
}

type salesPaymentRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_payment"`
}

func NewSalesPaymentRest(params salesPaymentRestParams) *SalesPaymentRest {
	rest := &SalesPaymentRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesPaymentRest struct {
	composable.CrudRestBase
}

type salesFulfillmentRequestRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_fulfillment_request"`
}

func NewSalesFulfillmentRequestRest(params salesFulfillmentRequestRestParams) *SalesFulfillmentRequestRest {
	rest := &SalesFulfillmentRequestRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesFulfillmentRequestRest struct {
	composable.CrudRestBase
}

type salesFulfillmentRequestLineRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_fulfillment_request_line"`
}

func NewSalesFulfillmentRequestLineRest(params salesFulfillmentRequestLineRestParams) *SalesFulfillmentRequestLineRest {
	rest := &SalesFulfillmentRequestLineRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesFulfillmentRequestLineRest struct {
	composable.CrudRestBase
}

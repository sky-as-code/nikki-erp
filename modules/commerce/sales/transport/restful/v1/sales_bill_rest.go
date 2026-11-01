package v1

import (
	"go.uber.org/dig"

	itBilling "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/billing"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

const (
	SalesBillEngineName                   = "dynengine_sales_bill"
	SalesBillLineEngineName               = "dynengine_sales_bill_line"
	SalesBillRelationEngineName           = "dynengine_sales_bill_relation"
	SalesPaymentEngineName                = "dynengine_sales_payment"
	SalesFulfillmentRequestEngineName     = "dynengine_sales_fulfillment_request"
	SalesFulfillmentRequestLineEngineName = "dynengine_sales_fulfillment_request_line"
)

type salesBillRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_sales_bill"`
}

func NewSalesBillRest(params salesBillRestParams) *SalesBillRest {
	rest := &SalesBillRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.billSvc = params.Engine.ApplicationService().(itBilling.SalesBillApplicationService)
	return rest
}

type SalesBillRest struct {
	dynamicresource.CrudRestBase
	billSvc itBilling.SalesBillApplicationService
}

type salesBillLineRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_sales_bill_line"`
}

func NewSalesBillLineRest(params salesBillLineRestParams) *SalesBillLineRest {
	rest := &SalesBillLineRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesBillLineRest struct {
	dynamicresource.CrudRestBase
}

type salesBillRelationRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_sales_bill_relation"`
}

func NewSalesBillRelationRest(params salesBillRelationRestParams) *SalesBillRelationRest {
	rest := &SalesBillRelationRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesBillRelationRest struct {
	dynamicresource.CrudRestBase
}

type salesPaymentRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_sales_payment"`
}

func NewSalesPaymentRest(params salesPaymentRestParams) *SalesPaymentRest {
	rest := &SalesPaymentRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesPaymentRest struct {
	dynamicresource.CrudRestBase
}

type salesFulfillmentRequestRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_sales_fulfillment_request"`
}

func NewSalesFulfillmentRequestRest(params salesFulfillmentRequestRestParams) *SalesFulfillmentRequestRest {
	rest := &SalesFulfillmentRequestRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesFulfillmentRequestRest struct {
	dynamicresource.CrudRestBase
}

type salesFulfillmentRequestLineRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_sales_fulfillment_request_line"`
}

func NewSalesFulfillmentRequestLineRest(params salesFulfillmentRequestLineRestParams) *SalesFulfillmentRequestLineRest {
	rest := &SalesFulfillmentRequestLineRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesFulfillmentRequestLineRest struct {
	dynamicresource.CrudRestBase
}

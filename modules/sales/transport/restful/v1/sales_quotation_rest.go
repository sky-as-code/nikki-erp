package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itQuotation "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/quotation"
)

const (
	SalesQuotationEngineName     = "dynengine_sales_quotation"
	SalesQuotationLineEngineName = "dynengine_sales_quotation_line"
)

type salesQuotationRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_quotation"`
}

func NewSalesQuotationRest(params salesQuotationRestParams) *SalesQuotationRest {
	rest := &SalesQuotationRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.quotationSvc = params.Engine.ApplicationService().(itQuotation.SalesQuotationApplicationService)
	return rest
}

type SalesQuotationRest struct {
	composable.CrudRestBase
	quotationSvc itQuotation.SalesQuotationApplicationService
}

type salesQuotationLineRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_quotation_line"`
}

func NewSalesQuotationLineRest(params salesQuotationLineRestParams) *SalesQuotationLineRest {
	rest := &SalesQuotationLineRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesQuotationLineRest struct {
	composable.CrudRestBase
}

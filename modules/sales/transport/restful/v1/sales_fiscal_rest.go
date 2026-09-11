package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itFiscal "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/fiscal"
)

const (
	SalesFiscalRequestEngineName = "dynengine_sales_fiscal_request"
	SalesBillingInstructionEngineName = "dynengine_sales_billing_instruction"
	SalesBillingIssuanceAttemptEngineName = "dynengine_sales_billing_issuance_attempt"
)

type salesFiscalRequestRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_fiscal_request"`
}

func NewSalesFiscalRequestRest(params salesFiscalRequestRestParams) *SalesFiscalRequestRest {
	rest := &SalesFiscalRequestRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.fiscalSvc = params.Engine.ApplicationService().(itFiscal.SalesFiscalRequestApplicationService)
	return rest
}

type SalesFiscalRequestRest struct {
	composable.CrudRestBase
	fiscalSvc itFiscal.SalesFiscalRequestApplicationService
}

type salesBillingInstructionRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_billing_instruction"`
}

func NewSalesBillingInstructionRest(params salesBillingInstructionRestParams) *SalesBillingInstructionRest {
	rest := &SalesBillingInstructionRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.billingSvc = params.Engine.ApplicationService().(itFiscal.SalesBillingInstructionApplicationService)
	return rest
}

type SalesBillingInstructionRest struct {
	composable.CrudRestBase
	billingSvc itFiscal.SalesBillingInstructionApplicationService
}

type salesBillingIssuanceAttemptRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_billing_issuance_attempt"`
}

func NewSalesBillingIssuanceAttemptRest(params salesBillingIssuanceAttemptRestParams) *SalesBillingIssuanceAttemptRest {
	rest := &SalesBillingIssuanceAttemptRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesBillingIssuanceAttemptRest struct {
	composable.CrudRestBase
}

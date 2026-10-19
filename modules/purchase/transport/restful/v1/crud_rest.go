package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itAgreement "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/agreement"
	itOrder "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/order"
)

// The container names the REST structs inject. A dig name tag must be a literal, so routes_test.go
// asserts each still equals composable.EngineDependencyName of its schema.
const (
	ConfigurationEngineName      = "dynengine_purchase_configuration"
	VendorProductPriceEngineName = "dynengine_purchase_vendor_product_price"
	SourcingGroupEngineName      = "dynengine_purchase_sourcing_group"
	AuditEventEngineName         = "dynengine_purchase_audit_event"
	AgreementLineEngineName      = "dynengine_purchase_agreement_line"
	PurchaseOrderLineEngineName  = "dynengine_purchase_order_line"
	AgreementEngineName          = "dynengine_purchase_agreement"
	PurchaseOrderEngineName      = "dynengine_purchase_order"
)

type configurationRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_purchase_configuration"`
}

func NewConfigurationRest(params configurationRestParams) *ConfigurationRest {
	rest := &ConfigurationRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type ConfigurationRest struct {
	composable.CrudRestBase
}

type vendorProductPriceRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_purchase_vendor_product_price"`
}

func NewVendorProductPriceRest(params vendorProductPriceRestParams) *VendorProductPriceRest {
	rest := &VendorProductPriceRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type VendorProductPriceRest struct {
	composable.CrudRestBase
}

type sourcingGroupRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_purchase_sourcing_group"`
}

func NewSourcingGroupRest(params sourcingGroupRestParams) *SourcingGroupRest {
	rest := &SourcingGroupRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SourcingGroupRest struct {
	composable.CrudRestBase
}

type auditEventRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_purchase_audit_event"`
}

func NewAuditEventRest(params auditEventRestParams) *AuditEventRest {
	rest := &AuditEventRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type AuditEventRest struct {
	composable.CrudRestBase
}

type agreementLineRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_purchase_agreement_line"`
}

func NewAgreementLineRest(params agreementLineRestParams) *AgreementLineRest {
	rest := &AgreementLineRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type AgreementLineRest struct {
	composable.CrudRestBase
}

type purchaseOrderLineRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_purchase_order_line"`
}

func NewPurchaseOrderLineRest(params purchaseOrderLineRestParams) *PurchaseOrderLineRest {
	rest := &PurchaseOrderLineRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type PurchaseOrderLineRest struct {
	composable.CrudRestBase
}

type agreementRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_purchase_agreement"`
}

func NewAgreementRest(params agreementRestParams) *AgreementRest {
	rest := &AgreementRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.agreementSvc = params.Engine.ApplicationService().(itAgreement.AgreementCrudApplicationService)
	return rest
}

type AgreementRest struct {
	composable.CrudRestBase
	agreementSvc itAgreement.AgreementCrudApplicationService
}

type purchaseOrderRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_purchase_order"`
}

func NewPurchaseOrderRest(params purchaseOrderRestParams) *PurchaseOrderRest {
	rest := &PurchaseOrderRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.orderSvc = params.Engine.ApplicationService().(itOrder.PurchaseOrderCrudApplicationService)
	return rest
}

type PurchaseOrderRest struct {
	composable.CrudRestBase
	orderSvc itOrder.PurchaseOrderCrudApplicationService
}

package app

import (
	itAgreement "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/agreement"
	itAudit "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/audit"
	itConfiguration "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/configuration"
	itOrder "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/order"
	itVendorPrice "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/vendorprice"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// The application services of the resources that add nothing to the built-in CRUD. The default
// resolves the org, asserts the permission and confines single-row actions to the org; the write
// rules live in the domain services.

func NewConfigurationApplicationService(
	base dynamicresource.CrudApplicationService,
) itConfiguration.ConfigurationCrudApplicationService {
	return &ConfigurationApplicationServiceImpl{CrudApplicationService: base}
}

type ConfigurationApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
}

func NewVendorProductPriceApplicationService(
	base dynamicresource.CrudApplicationService,
) itVendorPrice.VendorProductPriceCrudApplicationService {
	return &VendorProductPriceApplicationServiceImpl{CrudApplicationService: base}
}

type VendorProductPriceApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
}

func NewSourcingGroupApplicationService(
	base dynamicresource.CrudApplicationService,
) itOrder.SourcingGroupCrudApplicationService {
	return &SourcingGroupApplicationServiceImpl{CrudApplicationService: base}
}

type SourcingGroupApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
}

func NewAuditEventApplicationService(
	base dynamicresource.CrudApplicationService,
) itAudit.AuditEventCrudApplicationService {
	return &AuditEventApplicationServiceImpl{CrudApplicationService: base}
}

type AuditEventApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
}

func NewAgreementLineApplicationService(
	base dynamicresource.CrudApplicationService,
) itAgreement.AgreementLineCrudApplicationService {
	return &AgreementLineApplicationServiceImpl{CrudApplicationService: base}
}

type AgreementLineApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
}

func NewPurchaseOrderLineApplicationService(
	base dynamicresource.CrudApplicationService,
) itOrder.PurchaseOrderLineCrudApplicationService {
	return &PurchaseOrderLineApplicationServiceImpl{CrudApplicationService: base}
}

type PurchaseOrderLineApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
}

package app

import (
	itAgreement "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/agreement"
	itAudit "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/audit"
	itConfiguration "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/configuration"
	itOrder "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/order"
	itVendorPrice "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/vendorprice"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// The application services of the resources that add nothing to the built-in CRUD. The default
// resolves the org, asserts the permission and confines single-row actions to the org; the write
// rules live in the domain services.

func NewConfigurationApplicationService(
	base composable.CrudApplicationService,
) itConfiguration.ConfigurationCrudApplicationService {
	return &ConfigurationApplicationServiceImpl{CrudApplicationService: base}
}

type ConfigurationApplicationServiceImpl struct {
	composable.CrudApplicationService
}

func NewVendorProductPriceApplicationService(
	base composable.CrudApplicationService,
) itVendorPrice.VendorProductPriceCrudApplicationService {
	return &VendorProductPriceApplicationServiceImpl{CrudApplicationService: base}
}

type VendorProductPriceApplicationServiceImpl struct {
	composable.CrudApplicationService
}

func NewSourcingGroupApplicationService(
	base composable.CrudApplicationService,
) itOrder.SourcingGroupCrudApplicationService {
	return &SourcingGroupApplicationServiceImpl{CrudApplicationService: base}
}

type SourcingGroupApplicationServiceImpl struct {
	composable.CrudApplicationService
}

func NewAuditEventApplicationService(
	base composable.CrudApplicationService,
) itAudit.AuditEventCrudApplicationService {
	return &AuditEventApplicationServiceImpl{CrudApplicationService: base}
}

type AuditEventApplicationServiceImpl struct {
	composable.CrudApplicationService
}

func NewAgreementLineApplicationService(
	base composable.CrudApplicationService,
) itAgreement.AgreementLineCrudApplicationService {
	return &AgreementLineApplicationServiceImpl{CrudApplicationService: base}
}

type AgreementLineApplicationServiceImpl struct {
	composable.CrudApplicationService
}

func NewPurchaseOrderLineApplicationService(
	base composable.CrudApplicationService,
) itOrder.PurchaseOrderLineCrudApplicationService {
	return &PurchaseOrderLineApplicationServiceImpl{CrudApplicationService: base}
}

type PurchaseOrderLineApplicationServiceImpl struct {
	composable.CrudApplicationService
}

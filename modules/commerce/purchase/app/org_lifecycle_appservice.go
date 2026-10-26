package app

import (
	"github.com/sky-as-code/nikki-erp/modules/commerce/purchase/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/core/orglifecycle"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// orgCleanUpOrder lists every org-scoped resource children first, as their foreign keys require.
var orgCleanUpOrder = []string{
	models.AgreementLineSchemaName, models.AgreementSchemaName, models.AuditEventSchemaName,
	models.ConfigurationSchemaName, models.PurchaseOrderLineSchemaName, models.PurchaseOrderSchemaName,
	models.SourcingGroupSchemaName, models.VendorProductPriceSchemaName,
}

func NewOrgLifecycleAppService() orglifecycle.OrgLifecycleService {
	return &OrgLifecycleAppServiceImpl{OrgDataCleanup: composable.NewOrgDataCleanup(orgCleanUpOrder...)}
}

// OrgLifecycleAppServiceImpl writes no defaults: a new organization starts with no purchase orders, agreements and vendor prices.
// Cleanup deletes the organization's.
type OrgLifecycleAppServiceImpl struct {
	composable.OrgDataCleanup
}

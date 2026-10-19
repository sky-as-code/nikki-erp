package app

import (
	"github.com/sky-as-code/nikki-erp/modules/core/orglifecycle"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
)

// orgCleanUpOrder lists every org-scoped resource children first, as their foreign keys require.
var orgCleanUpOrder = []string{
	models.InvoiceLineSchemaName, models.InvoiceSchemaName, models.OrderSchemaName,
	models.TransactionSchemaName, models.PaymentProfileSchemaName,
}

func NewOrgLifecycleAppService() orglifecycle.OrgLifecycleService {
	return &OrgLifecycleAppServiceImpl{OrgDataCleanup: composable.NewOrgDataCleanup(orgCleanUpOrder...)}
}

// OrgLifecycleAppServiceImpl writes no defaults: a new organization starts with no invoices, payment orders and payment profiles.
// Cleanup deletes the organization's.
type OrgLifecycleAppServiceImpl struct {
	composable.OrgDataCleanup
}

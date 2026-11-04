package app

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/orglifecycle"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	modconstants "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/constants"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/services"
	pinvsettings "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/settings"
	itExt "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/external"
)

// orgCleanUpOrder lists every org-scoped resource children first, as their foreign keys require.
var orgCleanUpOrder = []string{
	models.InvoiceLineSchemaName, models.InvoiceSchemaName, models.OrderSchemaName,
	models.TransactionSchemaName, models.PaymentProfileSchemaName,
}

func NewOrgLifecycleAppService(
	sequences itExt.SequenceExtService, seeder itExt.OrgSettingsSeedExtService,
) orglifecycle.OrgLifecycleService {
	return &OrgLifecycleAppServiceImpl{
		OrgDataCleanup: dynamicresource.NewOrgDataCleanup(orgCleanUpOrder...),
		sequences:      sequences,
		seeder:         seeder,
	}
}

// OrgLifecycleAppServiceImpl gives a new organization its invoice number series and nothing else:
// it starts with no invoices, payment orders and payment profiles. Cleanup deletes the
// organization's; the sequence definition and the settings row go with Essential's and Settings'
// own cleanup.
type OrgLifecycleAppServiceImpl struct {
	dynamicresource.OrgDataCleanup

	sequences itExt.SequenceExtService
	seeder    itExt.OrgSettingsSeedExtService
}

// InitOrg registers the invoice series and records its id in the org's settings, the only place
// the definition is registered. Both steps are idempotent, so a re-run lifecycle points the setting
// at the same series.
func (this *OrgLifecycleAppServiceImpl) InitOrg(ctx corectx.Context, orgId model.Id) error {
	_, err := itExt.RegisterOrgSequences(ctx, this.sequences, this.seeder, modconstants.PaymentInvoiceModuleName,
		orgId, itExt.SequenceOrgRegistration{
			Setting: pinvsettings.OrgSettingInvoiceNumberSequenceId, Command: services.InvoiceSequenceCommand(orgId),
		})
	return err
}

package app

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	purconstants "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/constants"
	"github.com/sky-as-code/nikki-erp/modules/commerce/purchase/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/commerce/purchase/domain/services"
	itExt "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/external"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/orglifecycle"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// orgCleanUpOrder lists every org-scoped resource children first, as their foreign keys require.
var orgCleanUpOrder = []string{
	models.AgreementLineSchemaName, models.AgreementSchemaName, models.AuditEventSchemaName,
	models.ConfigurationSchemaName, models.PurchaseOrderLineSchemaName, models.PurchaseOrderSchemaName,
	models.SourcingGroupSchemaName, models.VendorProductPriceSchemaName,
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

// OrgLifecycleAppServiceImpl gives a new organization its order and agreement code series and
// nothing else: it starts with no purchase orders, agreements and vendor prices. Cleanup deletes
// the organization's; the sequence definitions and settings rows go with Essential's and
// Settings' own cleanup.
type OrgLifecycleAppServiceImpl struct {
	dynamicresource.OrgDataCleanup

	sequences itExt.SequenceExtService
	seeder    itExt.OrgSettingsSeedExtService
}

// InitOrg registers the code series and records their ids in the org's Purchase settings, the only
// place a Purchase definition is registered. Both steps are idempotent, so a re-run lifecycle
// points the settings at the same series.
func (this *OrgLifecycleAppServiceImpl) InitOrg(ctx corectx.Context, orgId model.Id) error {
	_, err := itExt.RegisterOrgSequences(ctx, this.sequences, this.seeder,
		purconstants.PurchaseModuleName, orgId, services.OrgSequenceRegistrations(orgId)...)
	return err
}

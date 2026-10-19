package app

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/orglifecycle"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itEnt "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/entitlement"
	itOrgUnit "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/orgunit"
)

func NewOrgLifecycleAppService(
	entitlementSvc itEnt.EntitlementCrudDomainService, orgUnitSvc itOrgUnit.OrgUnitCrudDomainService,
) orglifecycle.OrgLifecycleService {
	return &OrgLifecycleAppServiceImpl{entitlementSvc: entitlementSvc, orgUnitSvc: orgUnitSvc}
}

// OrgLifecycleAppServiceImpl is iam's own answer to the org lifecycle. The org's settings and its
// creator's membership are written with the org itself, so there is nothing more to set up.
type OrgLifecycleAppServiceImpl struct {
	entitlementSvc itEnt.EntitlementCrudDomainService
	orgUnitSvc     itOrgUnit.OrgUnitCrudDomainService
}

func (this *OrgLifecycleAppServiceImpl) InitOrg(corectx.Context, model.Id) error {
	return nil
}

// CleanUpOrg removes the org's entitlements, whose foreign key would refuse the org's removal,
// and its units. Memberships and the permission cache cascade with the org row.
func (this *OrgLifecycleAppServiceImpl) CleanUpOrg(ctx corectx.Context, orgId model.Id) error {
	return composable.InTransaction(ctx, this.entitlementSvc.Repository(), func(tranxCtx corectx.Context) error {
		return composable.DeleteOrgData(tranxCtx, orgId, this.entitlementSvc, this.orgUnitSvc)
	})
}

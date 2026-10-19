package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itOrgUnit "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/orgunit"
)

// NewOrgUnitCrudApplicationService is handed the composable default by the iam_orgunit onion.
func NewOrgUnitCrudApplicationService(
	base composable.CrudApplicationService, domSvc itOrgUnit.OrgUnitCrudDomainService,
) itOrgUnit.OrgUnitCrudApplicationService {
	return &OrgUnitCrudApplicationServiceImpl{CrudApplicationService: base, domSvc: domSvc}
}

type OrgUnitCrudApplicationServiceImpl struct {
	composable.CrudApplicationService
	domSvc itOrgUnit.OrgUnitCrudDomainService
}

func (this *OrgUnitCrudApplicationServiceImpl) ManageUsers(
	ctx corectx.Context, cmd itOrgUnit.ManageOrgUnitUsersCommand,
) (*itOrgUnit.ManageOrgUnitUsersResult, error) {
	cErrs, err := composable.AssertRecordAction(ctx, this, cmd, composable.ActionPermission{ActionCode: actionManageUsers})
	if err != nil || cErrs != nil {
		return mutateRefused(cErrs, err)
	}
	return this.domSvc.ManageUsers(ctx, cmd)
}

package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	itOrgUnit "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/orgunit"
)

// NewOrgUnitCrudApplicationService is handed the composable default by the iam_orgunit onion.
func NewOrgUnitCrudApplicationService(
	base dynamicresource.CrudApplicationService, domSvc itOrgUnit.OrgUnitCrudDomainService,
) itOrgUnit.OrgUnitCrudApplicationService {
	return &OrgUnitCrudApplicationServiceImpl{CrudApplicationService: base, domSvc: domSvc}
}

type OrgUnitCrudApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
	domSvc itOrgUnit.OrgUnitCrudDomainService
}

func (this *OrgUnitCrudApplicationServiceImpl) ManageUsers(
	ctx corectx.Context, cmd itOrgUnit.ManageOrgUnitUsersCommand,
) (*itOrgUnit.ManageOrgUnitUsersResult, error) {
	cErrs, err := dynamicresource.AssertRecordAction(
		ctx, this, cmd, dynamicresource.ActionPermission{ActionCode: actionManageUsers},
	)
	if err != nil || cErrs != nil {
		return mutateRefused(cErrs, err)
	}
	return this.domSvc.ManageUsers(ctx, cmd)
}

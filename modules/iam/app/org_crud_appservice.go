package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itOrg "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/organization"
)

// NewOrgCrudApplicationService is handed the composable default by the iam_org onion.
func NewOrgCrudApplicationService(
	base composable.CrudApplicationService, domSvc itOrg.OrgCrudDomainService,
) itOrg.OrgCrudApplicationService {
	return &OrgCrudApplicationServiceImpl{CrudApplicationService: base, domSvc: domSvc}
}

type OrgCrudApplicationServiceImpl struct {
	composable.CrudApplicationService
	domSvc itOrg.OrgCrudDomainService
}

func (this *OrgCrudApplicationServiceImpl) ManageUsers(
	ctx corectx.Context, cmd itOrg.ManageOrgUsersCommand,
) (*itOrg.ManageOrgUsersResult, error) {
	cErrs, err := composable.AssertRecordAction(ctx, this, cmd, composable.ActionPermission{ActionCode: actionManageUsers})
	if err != nil || cErrs != nil {
		return mutateRefused(cErrs, err)
	}
	return this.domSvc.ManageUsers(ctx, cmd)
}

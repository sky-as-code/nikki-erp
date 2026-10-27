package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itService "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/service"
)

// NewServiceCrudApplicationService is handed the composable default by the iam_service onion,
// which authorizes every built-in action on the iam_service resource.
func NewServiceCrudApplicationService(
	base composable.CrudApplicationService, domSvc itService.ServiceCrudDomainService,
) itService.ServiceCrudApplicationService {
	return &ServiceCrudApplicationServiceImpl{CrudApplicationService: base, domSvc: domSvc}
}

type ServiceCrudApplicationServiceImpl struct {
	composable.CrudApplicationService
	domSvc itService.ServiceCrudDomainService
}

// ManageRoles asserts manage_role_assignments on iam_service, the permission the role side asserts
// for iam_role/:id/manage_services, so neither side opens a path the other keeps shut.
func (this *ServiceCrudApplicationServiceImpl) ManageRoles(
	ctx corectx.Context, cmd itService.ManageServiceRolesCommand,
) (*itService.ManageServiceRolesResult, error) {
	cErrs, err := composable.AssertRecordAction(ctx, this, cmd,
		composable.ActionPermission{ActionCode: actionManageRoleAssignments})
	if err != nil || cErrs != nil {
		return mutateRefused(cErrs, err)
	}
	return this.domSvc.ManageRoles(ctx, cmd)
}

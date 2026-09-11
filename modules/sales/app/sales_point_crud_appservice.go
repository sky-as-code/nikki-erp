package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itCatalog "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/catalog"
)

// NewSalesPointCrudApplicationService wraps the composable default with the point lifecycle.
//
// Named CrudApplicationService for the same reason as the channel's: the module already has a
// SalesPointApplicationServiceImpl serving the cross-module registration port.
func NewSalesPointCrudApplicationService(
	base composable.CrudApplicationService,
) itCatalog.SalesPointApplicationService {
	return &SalesPointCrudApplicationServiceImpl{
		CrudApplicationService: base,
		pointSvc:               base.DomainService().(itCatalog.SalesPointDomainService),
	}
}

type SalesPointCrudApplicationServiceImpl struct {
	composable.CrudApplicationService
	pointSvc itCatalog.SalesPointDomainService
}

func (this *SalesPointCrudApplicationServiceImpl) Suspend(
	ctx corectx.Context, cmd itCatalog.SalesPointLifecycleCommand,
) (*itCatalog.SalesPointLifecycleResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionSuspend, cmd); cErrs != nil || err != nil {
		return mutateFailure(cErrs, err)
	}
	return this.pointSvc.Suspend(ctx, readStringParam(cmd, paramRecordId))
}

func (this *SalesPointCrudApplicationServiceImpl) Activate(
	ctx corectx.Context, cmd itCatalog.SalesPointLifecycleCommand,
) (*itCatalog.SalesPointLifecycleResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionActivate, cmd); cErrs != nil || err != nil {
		return mutateFailure(cErrs, err)
	}
	return this.pointSvc.Activate(ctx, readStringParam(cmd, paramRecordId))
}

// Archive and Unarchive share the built-in set_archived permission: the same power in reverse.
func (this *SalesPointCrudApplicationServiceImpl) Archive(
	ctx corectx.Context, cmd itCatalog.SalesPointLifecycleCommand,
) (*itCatalog.SalesPointLifecycleResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, composable.PermissionSetArchived, cmd); cErrs != nil || err != nil {
		return mutateFailure(cErrs, err)
	}
	return this.pointSvc.Archive(ctx, readStringParam(cmd, paramRecordId))
}

func (this *SalesPointCrudApplicationServiceImpl) Unarchive(
	ctx corectx.Context, cmd itCatalog.SalesPointLifecycleCommand,
) (*itCatalog.SalesPointLifecycleResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, composable.PermissionSetArchived, cmd); cErrs != nil || err != nil {
		return mutateFailure(cErrs, err)
	}
	return this.pointSvc.Unarchive(ctx, readStringParam(cmd, paramRecordId))
}

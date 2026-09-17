package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itCatalog "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/catalog"
	itChannel "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/channel"
)

// NewSalesPointCrudApplicationService wraps the composable default with the point lifecycle.
//
// Named CrudApplicationService for the same reason as the channel's: the module already has a
// SalesPointApplicationServiceImpl serving the cross-module registration port.
func NewSalesPointCrudApplicationService(
	base composable.CrudApplicationService,
	pointPayments itChannel.PointPaymentAppService,
) itCatalog.SalesPointApplicationService {
	return &SalesPointCrudApplicationServiceImpl{
		CrudApplicationService: base,
		pointSvc:               base.DomainService().(itCatalog.SalesPointDomainService),
		pointPayments:          pointPayments,
	}
}

type SalesPointCrudApplicationServiceImpl struct {
	composable.CrudApplicationService
	pointSvc      itCatalog.SalesPointDomainService
	pointPayments itChannel.PointPaymentAppService
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

func (this *SalesPointCrudApplicationServiceImpl) PaymentMethods(
	ctx corectx.Context, query itCatalog.PointPaymentMethodsQuery,
) (*dyn.OpResult[any], error) {
	result, err := this.pointPayments.ListPointPaymentMethods(ctx, itChannel.ListPointPaymentMethodsQuery{
		SalesPointId: readStringParam(query, paramRecordId),
		EnabledOnly:  readBoolParam(query, "enabled_only"),
	})
	if err != nil {
		return nil, err
	}
	if result.ClientErrors.Count() > 0 {
		return &dyn.OpResult[any]{ClientErrors: result.ClientErrors}, nil
	}
	return anyResult(result.Data), nil
}

func (this *SalesPointCrudApplicationServiceImpl) EnablePaymentMethod(
	ctx corectx.Context, cmd itCatalog.PointPaymentMethodCommand,
) (*itCatalog.PointPaymentMethodResult, error) {
	return this.mapPointPaymentMethod(ctx, cmd, this.pointPayments.EnablePointPaymentMethod)
}

func (this *SalesPointCrudApplicationServiceImpl) DisablePaymentMethod(
	ctx corectx.Context, cmd itCatalog.PointPaymentMethodCommand,
) (*itCatalog.PointPaymentMethodResult, error) {
	return this.mapPointPaymentMethod(ctx, cmd, this.pointPayments.DisablePointPaymentMethod)
}

func (this *SalesPointCrudApplicationServiceImpl) mapPointPaymentMethod(
	ctx corectx.Context,
	cmd itCatalog.PointPaymentMethodCommand,
	call func(corectx.Context, itChannel.PointPaymentMethodCommand) (*itChannel.PointPaymentMutationResult, error),
) (*itCatalog.PointPaymentMethodResult, error) {
	result, err := call(ctx, itChannel.PointPaymentMethodCommand{
		SalesPointId:     readStringParam(cmd, paramRecordId),
		PaymentMethodId:  readStringParam(cmd, paramPaymentMethodId),
		PaymentProfileId: readStringParam(cmd, paramPaymentProfileId),
	})
	if err != nil {
		return nil, err
	}
	if result.ClientErrors.Count() > 0 {
		return &itCatalog.PointPaymentMethodResult{ClientErrors: result.ClientErrors}, nil
	}
	return &itCatalog.PointPaymentMethodResult{HasData: true}, nil
}

package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itCatalog "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/catalog"
	itChannel "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/channel"
)

// NewSalesChannelCrudApplicationService wraps the composable default with the channel lifecycle.
//
// Named CrudApplicationService rather than ApplicationService because the module already has a
// SalesChannelApplicationServiceImpl: that one is the cross-module registration port other modules
// call to claim a channel, this one is the resource's own authorized CRUD surface.
func NewSalesChannelCrudApplicationService(
	base composable.CrudApplicationService,
	channelPayments itChannel.ChannelPaymentAppService,
) itCatalog.SalesChannelApplicationService {
	return &SalesChannelCrudApplicationServiceImpl{
		CrudApplicationService: base,
		channelSvc:             base.DomainService().(itCatalog.SalesChannelDomainService),
		channelPayments:        channelPayments,
	}
}

type SalesChannelCrudApplicationServiceImpl struct {
	composable.CrudApplicationService
	channelSvc      itCatalog.SalesChannelDomainService
	channelPayments itChannel.ChannelPaymentAppService
}

// Suspend stops new sales points and new orders while leaving history, returns, refunds and fiscal
// adjustments working. It is its own permission because it is a business decision, distinct from
// archiving, which retires the record itself.
func (this *SalesChannelCrudApplicationServiceImpl) Suspend(
	ctx corectx.Context, cmd itCatalog.ChannelLifecycleCommand,
) (*itCatalog.ChannelLifecycleResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionSuspend, cmd); cErrs != nil || err != nil {
		return mutateFailure(cErrs, err)
	}
	return this.channelSvc.Suspend(ctx, readStringParam(cmd, paramRecordId))
}

func (this *SalesChannelCrudApplicationServiceImpl) Activate(
	ctx corectx.Context, cmd itCatalog.ChannelLifecycleCommand,
) (*itCatalog.ChannelLifecycleResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionActivate, cmd); cErrs != nil || err != nil {
		return mutateFailure(cErrs, err)
	}
	return this.channelSvc.Activate(ctx, readStringParam(cmd, paramRecordId))
}

// Archive rides on the built-in set_archived permission: splitting them would let a role archive a
// channel through one route while being refused on the other.
func (this *SalesChannelCrudApplicationServiceImpl) Archive(
	ctx corectx.Context, cmd itCatalog.ChannelLifecycleCommand,
) (*itCatalog.ChannelLifecycleResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, composable.PermissionSetArchived, cmd); cErrs != nil || err != nil {
		return mutateFailure(cErrs, err)
	}
	return this.channelSvc.Archive(ctx, readStringParam(cmd, paramRecordId))
}

// Resolve answers the channel a caller named by code. Collection-level: there is no record to
// place in an org yet, so it asserts the action without the record check.
func (this *SalesChannelCrudApplicationServiceImpl) Resolve(
	ctx corectx.Context, query itCatalog.ResolveChannelQuery,
) (*dyn.OpResult[any], error) {
	if _, cErrs := this.AssertAction(ctx, composable.PermissionRead, query); cErrs != nil {
		return anyFailure(cErrs, nil)
	}
	result, err := this.channelSvc.ResolveByCode(ctx, readStringParam(query, paramCode))
	if err != nil {
		return nil, err
	}
	if result.ClientErrors.Count() > 0 {
		return &dyn.OpResult[any]{ClientErrors: result.ClientErrors}, nil
	}
	return anyResult(result.Data), nil
}

// The channel's payment-method routes.
//
// These delegate to ChannelPaymentAppService rather than reimplementing the mapping: it already
// owns the merge with the upstream catalogue and performs its own permission check, so a second
// implementation here would be a second place for that check to drift. This layer only binds the
// bound field map into the typed command the service expects.

func (this *SalesChannelCrudApplicationServiceImpl) PaymentMethods(
	ctx corectx.Context, query itCatalog.ChannelPaymentMethodsQuery,
) (*dyn.OpResult[any], error) {
	result, err := this.channelPayments.ListChannelPaymentMethods(ctx, itChannel.ListChannelPaymentMethodsQuery{
		SalesChannelId: readStringParam(query, paramRecordId),
		EnabledOnly:    readBoolParam(query, "enabled_only"),
	})
	if err != nil {
		return nil, err
	}
	if result.ClientErrors.Count() > 0 {
		return &dyn.OpResult[any]{ClientErrors: result.ClientErrors}, nil
	}
	return anyResult(result.Data), nil
}

func (this *SalesChannelCrudApplicationServiceImpl) EnablePaymentMethod(
	ctx corectx.Context, cmd itCatalog.ChannelPaymentMethodCommand,
) (*itCatalog.ChannelPaymentMethodResult, error) {
	return this.mapPaymentMethod(ctx, cmd, this.channelPayments.EnableChannelPaymentMethod)
}

func (this *SalesChannelCrudApplicationServiceImpl) DisablePaymentMethod(
	ctx corectx.Context, cmd itCatalog.ChannelPaymentMethodCommand,
) (*itCatalog.ChannelPaymentMethodResult, error) {
	return this.mapPaymentMethod(ctx, cmd, this.channelPayments.DisableChannelPaymentMethod)
}

func (this *SalesChannelCrudApplicationServiceImpl) mapPaymentMethod(
	ctx corectx.Context,
	cmd itCatalog.ChannelPaymentMethodCommand,
	call func(corectx.Context, itChannel.ChannelPaymentMethodCommand) (*itChannel.ChannelPaymentMutationResult, error),
) (*itCatalog.ChannelPaymentMethodResult, error) {
	result, err := call(ctx, itChannel.ChannelPaymentMethodCommand{
		SalesChannelId:  readStringParam(cmd, paramRecordId),
		PaymentMethodId: readStringParam(cmd, paramPaymentMethodId),
	})
	if err != nil {
		return nil, err
	}
	if result.ClientErrors.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: result.ClientErrors}, nil
	}
	return &dyn.OpResult[dyn.MutateResultData]{
		Data:    dyn.MutateResultData{AffectedCount: 1},
		HasData: true,
	}, nil
}

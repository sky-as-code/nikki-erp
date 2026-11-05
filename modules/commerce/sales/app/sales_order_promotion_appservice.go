package app

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"

	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/services"
	itOrder "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/order"
)

// The draft order's promotion actions (plan §3.2). Each answers the PromotionQuote as its data.

const (
	paramAcceptedCalculationReference = "accepted_calculation_reference"
	paramReplaceExisting              = "replace_existing"
	paramPromotionId                  = "promotion_id"
	paramRewardId                     = "reward_id"
	paramDecline                      = "decline"
)

func (this *SalesOrderCrudApplicationServiceImpl) ApplyVoucher(
	ctx corectx.Context, cmd itOrder.OrderActionCommand,
) (*dyn.OpResult[any], error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionApplyVoucher, cmd); cErrs != nil || err != nil {
		return anyFailure(cErrs, err)
	}
	return quoteAnswer(services.ApplyOrderVoucher(ctx, readStringParam(cmd, paramRecordId),
		readStringParam(cmd, paramVoucherCode), readBoolParam(cmd, paramReplaceExisting), this.promotionDeps(ctx)))
}

// RemoveVoucher shares apply_voucher's permission: detaching is the same power as attaching.
func (this *SalesOrderCrudApplicationServiceImpl) RemoveVoucher(
	ctx corectx.Context, cmd itOrder.OrderActionCommand,
) (*dyn.OpResult[any], error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionApplyVoucher, cmd); cErrs != nil || err != nil {
		return anyFailure(cErrs, err)
	}
	return quoteAnswer(services.RemoveOrderVoucher(ctx, readStringParam(cmd, paramRecordId),
		readStringParam(cmd, paramPromotionId), this.promotionDeps(ctx)))
}

func (this *SalesOrderCrudApplicationServiceImpl) SelectPromotionReward(
	ctx corectx.Context, cmd itOrder.OrderActionCommand,
) (*dyn.OpResult[any], error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionReprice, cmd); cErrs != nil || err != nil {
		return anyFailure(cErrs, err)
	}
	return quoteAnswer(services.SelectPromotionReward(ctx, readStringParam(cmd, paramRecordId),
		readStringParam(cmd, paramPromotionId), readStringParam(cmd, paramRewardId),
		readBoolParam(cmd, paramDecline), this.promotionDeps(ctx)))
}

// PreviewPromotions reprices and persists the draft; it consumes nothing.
func (this *SalesOrderCrudApplicationServiceImpl) PreviewPromotions(
	ctx corectx.Context, cmd itOrder.OrderActionCommand,
) (*dyn.OpResult[any], error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionReprice, cmd); cErrs != nil || err != nil {
		return anyFailure(cErrs, err)
	}
	return quoteAnswer(services.PreviewPromotions(ctx, readStringParam(cmd, paramRecordId), this.promotionDeps(ctx)))
}

func (this *SalesOrderCrudApplicationServiceImpl) promotionDeps(ctx corectx.Context) services.PromotionActionDeps {
	return services.PromotionActionDeps{
		TaxSvc: this.taxCalculation, Policy: services.ResolveSalesPolicy(ctx, this.effectiveSettings),
		BasisSvc: this.pricingBasis,
	}
}

func quoteAnswer(quote map[string]any, vErrs *ft.ClientErrors, err error) (*dyn.OpResult[any], error) {
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &dyn.OpResult[any]{ClientErrors: *vErrs}, nil
	}
	return &dyn.OpResult[any]{HasData: true, Data: quote}, nil
}

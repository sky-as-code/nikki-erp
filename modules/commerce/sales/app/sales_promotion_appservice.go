package app

import (
	"github.com/samber/lo"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itEvent "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/event"
	itPromotion "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/promotion"
)

// The promotion and voucher application services. No custom actions: a discount is defined here
// and applied by the order's apply_voucher, which lives on the order.

func NewSalesPromotionProgramApplicationService(
	base composable.CrudApplicationService, pricingChangedPub itEvent.PricingChangedEventPublisher,
) itPromotion.SalesPromotionProgramApplicationService {
	return &SalesPromotionProgramApplicationServiceImpl{CrudApplicationService: base, pricingChangedPub: pricingChangedPub}
}

type SalesPromotionProgramApplicationServiceImpl struct {
	composable.CrudApplicationService

	pricingChangedPub itEvent.PricingChangedEventPublisher
}

func NewSalesPromotionConditionGroupApplicationService(
	base composable.CrudApplicationService, pricingChangedPub itEvent.PricingChangedEventPublisher,
) itPromotion.SalesPromotionConditionGroupApplicationService {
	return &SalesPromotionConditionGroupApplicationServiceImpl{CrudApplicationService: base, pricingChangedPub: pricingChangedPub}
}

type SalesPromotionConditionGroupApplicationServiceImpl struct {
	composable.CrudApplicationService

	pricingChangedPub itEvent.PricingChangedEventPublisher
}

func NewSalesPromotionConditionApplicationService(
	base composable.CrudApplicationService, pricingChangedPub itEvent.PricingChangedEventPublisher,
) itPromotion.SalesPromotionConditionApplicationService {
	return &SalesPromotionConditionApplicationServiceImpl{CrudApplicationService: base, pricingChangedPub: pricingChangedPub}
}

type SalesPromotionConditionApplicationServiceImpl struct {
	composable.CrudApplicationService

	pricingChangedPub itEvent.PricingChangedEventPublisher
}

func NewSalesPromotionConditionTargetApplicationService(
	base composable.CrudApplicationService, pricingChangedPub itEvent.PricingChangedEventPublisher,
) itPromotion.SalesPromotionConditionTargetApplicationService {
	return &SalesPromotionConditionTargetApplicationServiceImpl{CrudApplicationService: base, pricingChangedPub: pricingChangedPub}
}

type SalesPromotionConditionTargetApplicationServiceImpl struct {
	composable.CrudApplicationService

	pricingChangedPub itEvent.PricingChangedEventPublisher
}

func NewSalesPromotionRewardApplicationService(
	base composable.CrudApplicationService, pricingChangedPub itEvent.PricingChangedEventPublisher,
) itPromotion.SalesPromotionRewardApplicationService {
	return &SalesPromotionRewardApplicationServiceImpl{CrudApplicationService: base, pricingChangedPub: pricingChangedPub}
}

type SalesPromotionRewardApplicationServiceImpl struct {
	composable.CrudApplicationService

	pricingChangedPub itEvent.PricingChangedEventPublisher
}

func NewSalesPromotionCompatibilityApplicationService(
	base composable.CrudApplicationService, pricingChangedPub itEvent.PricingChangedEventPublisher,
) itPromotion.SalesPromotionCompatibilityApplicationService {
	return &SalesPromotionCompatibilityApplicationServiceImpl{CrudApplicationService: base, pricingChangedPub: pricingChangedPub}
}

type SalesPromotionCompatibilityApplicationServiceImpl struct {
	composable.CrudApplicationService

	pricingChangedPub itEvent.PricingChangedEventPublisher
}

func NewSalesVoucherCodeApplicationService(
	base composable.CrudApplicationService, pricingChangedPub itEvent.PricingChangedEventPublisher,
) itPromotion.SalesVoucherCodeApplicationService {
	return &SalesVoucherCodeApplicationServiceImpl{CrudApplicationService: base, pricingChangedPub: pricingChangedPub}
}

type SalesVoucherCodeApplicationServiceImpl struct {
	composable.CrudApplicationService

	pricingChangedPub itEvent.PricingChangedEventPublisher
}

func NewSalesVoucherRedemptionApplicationService(base composable.CrudApplicationService) itPromotion.SalesVoucherRedemptionApplicationService {
	return &SalesVoucherRedemptionApplicationServiceImpl{CrudApplicationService: base}
}

type SalesVoucherRedemptionApplicationServiceImpl struct {
	composable.CrudApplicationService
}

func (this *SalesPromotionProgramApplicationServiceImpl) Create(
	ctx corectx.Context, cmd composable.CreateCommand,
) (*composable.CreateResult, error) {
	res, err := this.CrudApplicationService.Create(ctx, cmd)
	if err == nil && res.HasData && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedPromotionProgram, lo.FromPtr(res.Data.GetString(basemodel.FieldOrgId)), lo.FromPtr(res.Data.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesPromotionProgramApplicationServiceImpl) Update(
	ctx corectx.Context, cmd composable.UpdateCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.Update(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedPromotionProgram, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesPromotionProgramApplicationServiceImpl) Delete(
	ctx corectx.Context, cmd composable.DeleteCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.Delete(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedPromotionProgram, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesPromotionProgramApplicationServiceImpl) SetArchived(
	ctx corectx.Context, cmd composable.SetArchivedCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.SetArchived(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedPromotionProgram, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesPromotionConditionGroupApplicationServiceImpl) Create(
	ctx corectx.Context, cmd composable.CreateCommand,
) (*composable.CreateResult, error) {
	res, err := this.CrudApplicationService.Create(ctx, cmd)
	if err == nil && res.HasData && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedPromotionConditionGroup, lo.FromPtr(res.Data.GetString(basemodel.FieldOrgId)), lo.FromPtr(res.Data.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesPromotionConditionGroupApplicationServiceImpl) Update(
	ctx corectx.Context, cmd composable.UpdateCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.Update(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedPromotionConditionGroup, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesPromotionConditionGroupApplicationServiceImpl) Delete(
	ctx corectx.Context, cmd composable.DeleteCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.Delete(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedPromotionConditionGroup, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesPromotionConditionGroupApplicationServiceImpl) SetArchived(
	ctx corectx.Context, cmd composable.SetArchivedCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.SetArchived(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedPromotionConditionGroup, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesPromotionConditionApplicationServiceImpl) Create(
	ctx corectx.Context, cmd composable.CreateCommand,
) (*composable.CreateResult, error) {
	res, err := this.CrudApplicationService.Create(ctx, cmd)
	if err == nil && res.HasData && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedPromotionCondition, lo.FromPtr(res.Data.GetString(basemodel.FieldOrgId)), lo.FromPtr(res.Data.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesPromotionConditionApplicationServiceImpl) Update(
	ctx corectx.Context, cmd composable.UpdateCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.Update(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedPromotionCondition, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesPromotionConditionApplicationServiceImpl) Delete(
	ctx corectx.Context, cmd composable.DeleteCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.Delete(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedPromotionCondition, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesPromotionConditionApplicationServiceImpl) SetArchived(
	ctx corectx.Context, cmd composable.SetArchivedCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.SetArchived(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedPromotionCondition, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesPromotionConditionTargetApplicationServiceImpl) Create(
	ctx corectx.Context, cmd composable.CreateCommand,
) (*composable.CreateResult, error) {
	res, err := this.CrudApplicationService.Create(ctx, cmd)
	if err == nil && res.HasData && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedPromotionConditionTarget, lo.FromPtr(res.Data.GetString(basemodel.FieldOrgId)), lo.FromPtr(res.Data.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesPromotionConditionTargetApplicationServiceImpl) Update(
	ctx corectx.Context, cmd composable.UpdateCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.Update(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedPromotionConditionTarget, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesPromotionConditionTargetApplicationServiceImpl) Delete(
	ctx corectx.Context, cmd composable.DeleteCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.Delete(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedPromotionConditionTarget, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesPromotionConditionTargetApplicationServiceImpl) SetArchived(
	ctx corectx.Context, cmd composable.SetArchivedCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.SetArchived(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedPromotionConditionTarget, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesPromotionRewardApplicationServiceImpl) Create(
	ctx corectx.Context, cmd composable.CreateCommand,
) (*composable.CreateResult, error) {
	res, err := this.CrudApplicationService.Create(ctx, cmd)
	if err == nil && res.HasData && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedPromotionReward, lo.FromPtr(res.Data.GetString(basemodel.FieldOrgId)), lo.FromPtr(res.Data.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesPromotionRewardApplicationServiceImpl) Update(
	ctx corectx.Context, cmd composable.UpdateCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.Update(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedPromotionReward, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesPromotionRewardApplicationServiceImpl) Delete(
	ctx corectx.Context, cmd composable.DeleteCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.Delete(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedPromotionReward, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesPromotionRewardApplicationServiceImpl) SetArchived(
	ctx corectx.Context, cmd composable.SetArchivedCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.SetArchived(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedPromotionReward, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesPromotionCompatibilityApplicationServiceImpl) Create(
	ctx corectx.Context, cmd composable.CreateCommand,
) (*composable.CreateResult, error) {
	res, err := this.CrudApplicationService.Create(ctx, cmd)
	if err == nil && res.HasData && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedPromotionCompatibility, lo.FromPtr(res.Data.GetString(basemodel.FieldOrgId)), lo.FromPtr(res.Data.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesPromotionCompatibilityApplicationServiceImpl) Update(
	ctx corectx.Context, cmd composable.UpdateCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.Update(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedPromotionCompatibility, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesPromotionCompatibilityApplicationServiceImpl) Delete(
	ctx corectx.Context, cmd composable.DeleteCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.Delete(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedPromotionCompatibility, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesPromotionCompatibilityApplicationServiceImpl) SetArchived(
	ctx corectx.Context, cmd composable.SetArchivedCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.SetArchived(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedPromotionCompatibility, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesVoucherCodeApplicationServiceImpl) Create(
	ctx corectx.Context, cmd composable.CreateCommand,
) (*composable.CreateResult, error) {
	res, err := this.CrudApplicationService.Create(ctx, cmd)
	if err == nil && res.HasData && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedVoucherCode, lo.FromPtr(res.Data.GetString(basemodel.FieldOrgId)), lo.FromPtr(res.Data.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesVoucherCodeApplicationServiceImpl) Update(
	ctx corectx.Context, cmd composable.UpdateCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.Update(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedVoucherCode, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesVoucherCodeApplicationServiceImpl) Delete(
	ctx corectx.Context, cmd composable.DeleteCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.Delete(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedVoucherCode, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *SalesVoucherCodeApplicationServiceImpl) SetArchived(
	ctx corectx.Context, cmd composable.SetArchivedCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.SetArchived(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.pricingChangedPub.PublishAsync(ctx, itEvent.PricingChangedVoucherCode, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

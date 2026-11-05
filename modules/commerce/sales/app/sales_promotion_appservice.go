package app

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	c "github.com/sky-as-code/nikki-erp/modules/commerce/sales/constants"
	itPromotion "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/promotion"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// The promotion application services (BR §9.1, §9.4). Every action asserts its own permission on
// the promotion; a child rule or reward also asserts update on the promotion it belongs to, so the
// right to process orders never implies the right to change a promotion.

func NewSalesPromotionApplicationService(
	base dynamicresource.CrudApplicationService,
) itPromotion.SalesPromotionApplicationService {
	return &SalesPromotionApplicationServiceImpl{
		CrudApplicationService: base,
		promotionSvc:           base.DomainService().(itPromotion.SalesPromotionDomainService),
	}
}

type SalesPromotionApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
	promotionSvc itPromotion.SalesPromotionDomainService
}

func (this *SalesPromotionApplicationServiceImpl) Activate(
	ctx corectx.Context, cmd itPromotion.LifecycleCommand,
) (*itPromotion.RecordResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionActivate, cmd); cErrs != nil || err != nil {
		return recordFailure(cErrs, err)
	}
	return this.promotionSvc.Activate(ctx, readStringParam(cmd, paramRecordId))
}

func (this *SalesPromotionApplicationServiceImpl) Suspend(
	ctx corectx.Context, cmd itPromotion.LifecycleCommand,
) (*itPromotion.RecordResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionSuspend, cmd); cErrs != nil || err != nil {
		return recordFailure(cErrs, err)
	}
	return this.promotionSvc.Suspend(ctx, readStringParam(cmd, paramRecordId))
}

func (this *SalesPromotionApplicationServiceImpl) Duplicate(
	ctx corectx.Context, cmd itPromotion.LifecycleCommand,
) (*itPromotion.RecordResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionDuplicatePromotion, cmd); cErrs != nil || err != nil {
		return recordFailure(cErrs, err)
	}
	return this.promotionSvc.Duplicate(ctx, readStringParam(cmd, paramRecordId))
}

func (this *SalesPromotionApplicationServiceImpl) ManageChannels(
	ctx corectx.Context, cmd itPromotion.ManageCommand,
) (*itPromotion.ManageResult, error) {
	return this.manageScope(ctx, itPromotion.PromotionScopeChannels, cmd)
}

func (this *SalesPromotionApplicationServiceImpl) ManagePoints(
	ctx corectx.Context, cmd itPromotion.ManageCommand,
) (*itPromotion.ManageResult, error) {
	return this.manageScope(ctx, itPromotion.PromotionScopePoints, cmd)
}

func (this *SalesPromotionApplicationServiceImpl) ManagePricelists(
	ctx corectx.Context, cmd itPromotion.ManageCommand,
) (*itPromotion.ManageResult, error) {
	return this.manageScope(ctx, itPromotion.PromotionScopePricelists, cmd)
}

func (this *SalesPromotionApplicationServiceImpl) manageScope(
	ctx corectx.Context, scope itPromotion.PromotionScope, cmd itPromotion.ManageCommand,
) (*itPromotion.ManageResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionManagePromotion, cmd); cErrs != nil || err != nil {
		return mutateFailure(cErrs, err)
	}
	return this.promotionSvc.ManageScope(ctx, scope, cmd)
}

func recordFailure(cErrs *ft.ClientErrors, err error) (*itPromotion.RecordResult, error) {
	if err != nil {
		return nil, err
	}
	return &itPromotion.RecordResult{ClientErrors: *cErrs}, nil
}

// promotionChildAppService is the authorized surface of a rule or a reward. The built-in CRUD
// asserts the child's own permission; each write here also asserts update on the promotion.
type promotionChildAppService struct {
	dynamicresource.CrudApplicationService
}

func (this *promotionChildAppService) Create(
	ctx corectx.Context, cmd dynamicresource.CreateCommand,
) (*dynamicresource.CreateResult, error) {
	orgId, cErrs := this.AssertAction(ctx, dynamicresource.PermissionCreate, cmd)
	if cErrs == nil {
		cErrs = assertPromotionUpdate(ctx, orgId)
	}
	if cErrs != nil {
		return &dynamicresource.CreateResult{ClientErrors: *cErrs}, nil
	}
	return this.CrudApplicationService.Create(ctx, cmd)
}

func (this *promotionChildAppService) Update(
	ctx corectx.Context, cmd dynamicresource.UpdateCommand,
) (*dynamicresource.MutateResult, error) {
	if cErrs, err := this.assertParentUpdate(ctx, dynamicresource.PermissionUpdate, cmd); cErrs != nil || err != nil {
		return mutateFailure(cErrs, err)
	}
	return this.CrudApplicationService.Update(ctx, cmd)
}

func (this *promotionChildAppService) Delete(
	ctx corectx.Context, cmd dynamicresource.DeleteCommand,
) (*dynamicresource.MutateResult, error) {
	if cErrs, err := this.assertParentUpdate(ctx, dynamicresource.PermissionDelete, cmd); cErrs != nil || err != nil {
		return mutateFailure(cErrs, err)
	}
	return this.CrudApplicationService.Delete(ctx, cmd)
}

// assertParentUpdate asserts the child's own action on the record, then update on the promotion.
func (this *promotionChildAppService) assertParentUpdate(
	ctx corectx.Context, action string, cmd dmodel.DynamicFields,
) (*ft.ClientErrors, error) {
	orgId, cErrs := this.AssertAction(ctx, action, cmd)
	if cErrs != nil {
		return cErrs, nil
	}
	if orgId != nil {
		if cErrs, err := this.AssertRecordInOrg(ctx, cmd, *orgId); cErrs != nil || err != nil {
			return cErrs, err
		}
	}
	return assertPromotionUpdate(ctx, orgId), nil
}

func assertPromotionUpdate(ctx corectx.Context, orgId *model.Id) *ft.ClientErrors {
	if orgId == nil {
		return assertPermission(ctx, PermissionManagePromotion, c.SalesPromotionResource, c.ResourceScopeOrg)
	}
	return assertPermissionInOrg(ctx, PermissionManagePromotion, c.SalesPromotionResource,
		c.ResourceScopeOrg, *orgId)
}

type selectorManager interface {
	ManageProducts(ctx corectx.Context, cmd itPromotion.ManageCommand) (*itPromotion.ManageResult, error)
	ManageCategories(ctx corectx.Context, cmd itPromotion.ManageCommand) (*itPromotion.ManageResult, error)
}

// promotionSelectorOwnerAppService adds the manage_products and manage_categories actions.
type promotionSelectorOwnerAppService struct {
	promotionChildAppService
	selectors selectorManager
}

func (this *promotionSelectorOwnerAppService) ManageProducts(
	ctx corectx.Context, cmd itPromotion.ManageCommand,
) (*itPromotion.ManageResult, error) {
	if cErrs, err := this.assertParentUpdate(ctx, dynamicresource.PermissionUpdate, cmd); cErrs != nil || err != nil {
		return mutateFailure(cErrs, err)
	}
	return this.selectors.ManageProducts(ctx, cmd)
}

func (this *promotionSelectorOwnerAppService) ManageCategories(
	ctx corectx.Context, cmd itPromotion.ManageCommand,
) (*itPromotion.ManageResult, error) {
	if cErrs, err := this.assertParentUpdate(ctx, dynamicresource.PermissionUpdate, cmd); cErrs != nil || err != nil {
		return mutateFailure(cErrs, err)
	}
	return this.selectors.ManageCategories(ctx, cmd)
}

func NewSalesPromotionRuleApplicationService(
	base dynamicresource.CrudApplicationService,
) itPromotion.SalesPromotionRuleApplicationService {
	return &promotionSelectorOwnerAppService{
		promotionChildAppService: promotionChildAppService{CrudApplicationService: base},
		selectors:                base.DomainService().(itPromotion.SalesPromotionRuleDomainService),
	}
}

func NewSalesPromotionRewardApplicationService(
	base dynamicresource.CrudApplicationService,
) itPromotion.SalesPromotionRewardApplicationService {
	return &promotionSelectorOwnerAppService{
		promotionChildAppService: promotionChildAppService{CrudApplicationService: base},
		selectors:                base.DomainService().(itPromotion.SalesPromotionRewardDomainService),
	}
}

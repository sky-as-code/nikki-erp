package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	itPromotion "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/promotion"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// The promotion REST surface (plan §3.2). Lifecycle actions answer the record; manage_* answer the
// mutation like the built-in update.

type salesPromotionRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_sales_promotion"`
}

func NewSalesPromotionRest(params salesPromotionRestParams) *SalesPromotionRest {
	rest := &SalesPromotionRest{
		promotionSvc: params.Engine.ApplicationService().(itPromotion.SalesPromotionApplicationService),
	}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesPromotionRest struct {
	dynamicresource.CrudRestBase
	promotionSvc itPromotion.SalesPromotionApplicationService
}

func (this *SalesPromotionRest) Activate(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "activate sales promotion", payload,
		this.promotionSvc.Activate, dynamicresource.Identity[dmodel.DynamicFields])
}

func (this *SalesPromotionRest) Suspend(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "suspend sales promotion", payload,
		this.promotionSvc.Suspend, dynamicresource.Identity[dmodel.DynamicFields])
}

func (this *SalesPromotionRest) Duplicate(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "duplicate sales promotion", payload,
		this.promotionSvc.Duplicate, dynamicresource.Identity[dmodel.DynamicFields])
}

func (this *SalesPromotionRest) ManageChannels(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "manage promotion channels", payload,
		this.promotionSvc.ManageChannels, manageResponse)
}

func (this *SalesPromotionRest) ManagePoints(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "manage promotion points", payload,
		this.promotionSvc.ManagePoints, manageResponse)
}

func (this *SalesPromotionRest) ManagePricelists(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "manage promotion pricelists", payload,
		this.promotionSvc.ManagePricelists, manageResponse)
}

func manageResponse(mutation dyn.MutateResultData) any {
	return dynamicresource.MutateResponse(mutation)
}

// selectorOwnerService is the manage surface of a rule or a reward.
type selectorOwnerService interface {
	ManageProducts(ctx corectx.Context, cmd itPromotion.ManageCommand) (*itPromotion.ManageResult, error)
	ManageCategories(ctx corectx.Context, cmd itPromotion.ManageCommand) (*itPromotion.ManageResult, error)
}

// SalesPromotionSelectorOwnerRest serves a rule or a reward: CRUD plus manage_products and
// manage_categories.
type SalesPromotionSelectorOwnerRest struct {
	dynamicresource.CrudRestBase
	ownerSvc selectorOwnerService
	label    string
}

func newSelectorOwnerRest(engine dynamicresource.DynamicResourceEngineOnion, label string) *SalesPromotionSelectorOwnerRest {
	rest := &SalesPromotionSelectorOwnerRest{
		ownerSvc: engine.ApplicationService().(selectorOwnerService), label: label,
	}
	rest.SetApplicationService(engine.ApplicationService())
	return rest
}

func (this *SalesPromotionSelectorOwnerRest) ManageProducts(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "manage "+this.label+" products", payload,
		this.ownerSvc.ManageProducts, manageResponse)
}

func (this *SalesPromotionSelectorOwnerRest) ManageCategories(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "manage "+this.label+" categories", payload,
		this.ownerSvc.ManageCategories, manageResponse)
}

type salesPromotionRuleRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_sales_promotion_rule"`
}

// SalesPromotionRuleRest is the rule's REST handler set.
type SalesPromotionRuleRest struct {
	*SalesPromotionSelectorOwnerRest
}

func NewSalesPromotionRuleRest(params salesPromotionRuleRestParams) *SalesPromotionRuleRest {
	return &SalesPromotionRuleRest{newSelectorOwnerRest(params.Engine, "promotion rule")}
}

type salesPromotionRewardRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_sales_promotion_reward"`
}

// SalesPromotionRewardRest is the reward's REST handler set.
type SalesPromotionRewardRest struct {
	*SalesPromotionSelectorOwnerRest
}

func NewSalesPromotionRewardRest(params salesPromotionRewardRestParams) *SalesPromotionRewardRest {
	return &SalesPromotionRewardRest{newSelectorOwnerRest(params.Engine, "promotion reward")}
}

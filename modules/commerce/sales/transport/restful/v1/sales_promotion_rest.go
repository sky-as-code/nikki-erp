package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

const (
	SalesPromotionProgramEngineName = "dynengine_sales_promotion_program"
	SalesPromotionConditionGroupEngineName = "dynengine_sales_promotion_condition_group"
	SalesPromotionConditionEngineName = "dynengine_sales_promotion_condition"
	SalesPromotionConditionTargetEngineName = "dynengine_sales_promotion_condition_target"
	SalesPromotionRewardEngineName = "dynengine_sales_promotion_reward"
	SalesPromotionCompatibilityEngineName = "dynengine_sales_promotion_compatibility"
	SalesVoucherCodeEngineName = "dynengine_sales_voucher_code"
	SalesVoucherRedemptionEngineName = "dynengine_sales_voucher_redemption"
)

type salesPromotionProgramRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_promotion_program"`
}

func NewSalesPromotionProgramRest(params salesPromotionProgramRestParams) *SalesPromotionProgramRest {
	rest := &SalesPromotionProgramRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesPromotionProgramRest struct {
	composable.CrudRestBase
}

type salesPromotionConditionGroupRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_promotion_condition_group"`
}

func NewSalesPromotionConditionGroupRest(params salesPromotionConditionGroupRestParams) *SalesPromotionConditionGroupRest {
	rest := &SalesPromotionConditionGroupRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesPromotionConditionGroupRest struct {
	composable.CrudRestBase
}

type salesPromotionConditionRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_promotion_condition"`
}

func NewSalesPromotionConditionRest(params salesPromotionConditionRestParams) *SalesPromotionConditionRest {
	rest := &SalesPromotionConditionRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesPromotionConditionRest struct {
	composable.CrudRestBase
}

type salesPromotionConditionTargetRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_promotion_condition_target"`
}

func NewSalesPromotionConditionTargetRest(params salesPromotionConditionTargetRestParams) *SalesPromotionConditionTargetRest {
	rest := &SalesPromotionConditionTargetRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesPromotionConditionTargetRest struct {
	composable.CrudRestBase
}

type salesPromotionRewardRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_promotion_reward"`
}

func NewSalesPromotionRewardRest(params salesPromotionRewardRestParams) *SalesPromotionRewardRest {
	rest := &SalesPromotionRewardRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesPromotionRewardRest struct {
	composable.CrudRestBase
}

type salesPromotionCompatibilityRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_promotion_compatibility"`
}

func NewSalesPromotionCompatibilityRest(params salesPromotionCompatibilityRestParams) *SalesPromotionCompatibilityRest {
	rest := &SalesPromotionCompatibilityRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesPromotionCompatibilityRest struct {
	composable.CrudRestBase
}

type salesVoucherCodeRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_voucher_code"`
}

func NewSalesVoucherCodeRest(params salesVoucherCodeRestParams) *SalesVoucherCodeRest {
	rest := &SalesVoucherCodeRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesVoucherCodeRest struct {
	composable.CrudRestBase
}

type salesVoucherRedemptionRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_voucher_redemption"`
}

func NewSalesVoucherRedemptionRest(params salesVoucherRedemptionRestParams) *SalesVoucherRedemptionRest {
	rest := &SalesVoucherRedemptionRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesVoucherRedemptionRest struct {
	composable.CrudRestBase
}

package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

const (
	SalesPromotionProgramEngineName         = "dynengine_sales_promotion_program"
	SalesPromotionConditionGroupEngineName  = "dynengine_sales_promotion_condition_group"
	SalesPromotionConditionEngineName       = "dynengine_sales_promotion_condition"
	SalesPromotionConditionTargetEngineName = "dynengine_sales_promotion_condition_target"
	SalesPromotionRewardEngineName          = "dynengine_sales_promotion_reward"
	SalesPromotionCompatibilityEngineName   = "dynengine_sales_promotion_compatibility"
	SalesVoucherCodeEngineName              = "dynengine_sales_voucher_code"
	SalesVoucherRedemptionEngineName        = "dynengine_sales_voucher_redemption"
)

type salesPromotionProgramRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_sales_promotion_program"`
}

func NewSalesPromotionProgramRest(params salesPromotionProgramRestParams) *SalesPromotionProgramRest {
	rest := &SalesPromotionProgramRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesPromotionProgramRest struct {
	dynamicresource.CrudRestBase
}

type salesPromotionConditionGroupRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_sales_promotion_condition_group"`
}

func NewSalesPromotionConditionGroupRest(params salesPromotionConditionGroupRestParams) *SalesPromotionConditionGroupRest {
	rest := &SalesPromotionConditionGroupRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesPromotionConditionGroupRest struct {
	dynamicresource.CrudRestBase
}

type salesPromotionConditionRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_sales_promotion_condition"`
}

func NewSalesPromotionConditionRest(params salesPromotionConditionRestParams) *SalesPromotionConditionRest {
	rest := &SalesPromotionConditionRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesPromotionConditionRest struct {
	dynamicresource.CrudRestBase
}

type salesPromotionConditionTargetRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_sales_promotion_condition_target"`
}

func NewSalesPromotionConditionTargetRest(params salesPromotionConditionTargetRestParams) *SalesPromotionConditionTargetRest {
	rest := &SalesPromotionConditionTargetRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesPromotionConditionTargetRest struct {
	dynamicresource.CrudRestBase
}

type salesPromotionRewardRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_sales_promotion_reward"`
}

func NewSalesPromotionRewardRest(params salesPromotionRewardRestParams) *SalesPromotionRewardRest {
	rest := &SalesPromotionRewardRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesPromotionRewardRest struct {
	dynamicresource.CrudRestBase
}

type salesPromotionCompatibilityRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_sales_promotion_compatibility"`
}

func NewSalesPromotionCompatibilityRest(params salesPromotionCompatibilityRestParams) *SalesPromotionCompatibilityRest {
	rest := &SalesPromotionCompatibilityRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesPromotionCompatibilityRest struct {
	dynamicresource.CrudRestBase
}

type salesVoucherCodeRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_sales_voucher_code"`
}

func NewSalesVoucherCodeRest(params salesVoucherCodeRestParams) *SalesVoucherCodeRest {
	rest := &SalesVoucherCodeRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesVoucherCodeRest struct {
	dynamicresource.CrudRestBase
}

type salesVoucherRedemptionRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_sales_voucher_redemption"`
}

func NewSalesVoucherRedemptionRest(params salesVoucherRedemptionRestParams) *SalesVoucherRedemptionRest {
	rest := &SalesVoucherRedemptionRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type SalesVoucherRedemptionRest struct {
	dynamicresource.CrudRestBase
}

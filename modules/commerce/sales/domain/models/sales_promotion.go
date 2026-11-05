package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// The promotion configuration (BR §8.2-§8.6): the promotion itself, its channel/point/pricelist
// scope, its rules and rewards, and the product and category selectors those narrow by. Every
// child is configuration of the promotion: a write to one bumps the promotion's config_version and
// is refused while the promotion is active (BR-L02).

const (
	SalesPromotionSchemaName = "sales_promotion"

	SalesPromotionFieldId                      = basemodel.FieldId
	SalesPromotionFieldOrgId                   = basemodel.FieldOrgId
	SalesPromotionFieldReference               = "reference"
	SalesPromotionFieldName                    = "name"
	SalesPromotionFieldDescription             = "description"
	SalesPromotionFieldProgramType             = "program_type"
	SalesPromotionFieldStatus                  = "status"
	SalesPromotionFieldCurrencyId              = "currency_id"
	SalesPromotionFieldStartsAt                = "starts_at"
	SalesPromotionFieldEndsAt                  = "ends_at"
	SalesPromotionFieldUsageLimit              = "usage_limit"
	SalesPromotionFieldUsedCount               = "used_count"
	SalesPromotionFieldSequence                = "sequence"
	SalesPromotionFieldNextVoucherValidityDays = "next_voucher_validity_days"
	SalesPromotionFieldConfigVersion           = "config_version"

	SalesPromotionEdgeSalesChannels   = "sales_channels"
	SalesPromotionEdgeSalesPoints     = "sales_points"
	SalesPromotionEdgeSalesPricelists = "sales_pricelists"
)

// The scope junctions. Written only through the promotion's manage_* actions (crud.ManageM2m over
// the many:many edges above); they carry no org_id because the promotion they hang off is org-scoped.
const (
	SalesPromotionChannelSchemaName       = "sales_promotion_channel"
	SalesPromotionChannelFieldPromotionId = "promotion_id"
	SalesPromotionChannelFieldChannelId   = "sales_channel_id"

	SalesPromotionPointSchemaName       = "sales_promotion_point"
	SalesPromotionPointFieldPromotionId = "promotion_id"
	SalesPromotionPointFieldPointId     = "sales_point_id"

	SalesPromotionPricelistSchemaName       = "sales_promotion_pricelist"
	SalesPromotionPricelistFieldPromotionId = "promotion_id"
	SalesPromotionPricelistFieldPricelistId = "sales_pricelist_id"
)

const (
	SalesPromotionRuleSchemaName = "sales_promotion_rule"

	SalesPromotionRuleFieldId                 = basemodel.FieldId
	SalesPromotionRuleFieldPromotionId        = "promotion_id"
	SalesPromotionRuleFieldPhase              = "phase"
	SalesPromotionRuleFieldName               = "name"
	SalesPromotionRuleFieldMinQuantity        = "min_quantity"
	SalesPromotionRuleFieldQualificationUomId = "qualification_uom_id"
	SalesPromotionRuleFieldMinAmount          = "min_amount"
	SalesPromotionRuleFieldAmountTaxBasis     = "amount_tax_basis"
)

const (
	SalesPromotionRewardSchemaName = "sales_promotion_reward"

	SalesPromotionRewardFieldId                   = basemodel.FieldId
	SalesPromotionRewardFieldPromotionId          = "promotion_id"
	SalesPromotionRewardFieldRewardType           = "reward_type"
	SalesPromotionRewardFieldOrderDescription     = "order_description"
	SalesPromotionRewardFieldDiscountMethod       = "discount_method"
	SalesPromotionRewardFieldDiscountValue        = "discount_value"
	SalesPromotionRewardFieldDiscountScope        = "discount_scope"
	SalesPromotionRewardFieldMaxDiscountAmount    = "max_discount_amount"
	SalesPromotionRewardFieldGiftProductVariantId = "gift_product_variant_id"
	SalesPromotionRewardFieldGiftQuantity         = "gift_quantity"
	SalesPromotionRewardFieldGiftUomId            = "gift_uom_id"
)

// The selectors. A product selector names a template or a variant, never both; a category selector
// matches the category and its descendants (BR-C03).
const (
	SalesPromotionRuleProductSchemaName    = "sales_promotion_rule_product"
	SalesPromotionRuleCategorySchemaName   = "sales_promotion_rule_category"
	SalesPromotionRewardProductSchemaName  = "sales_promotion_reward_product"
	SalesPromotionRewardCategorySchemaName = "sales_promotion_reward_category"
	SalesPromotionSelectorFieldRuleId      = "rule_id"
	SalesPromotionSelectorFieldRewardId    = "reward_id"
	SalesPromotionSelectorFieldTemplateId  = "product_template_id"
	SalesPromotionSelectorFieldVariantId   = "product_variant_id"
	SalesPromotionSelectorFieldCategoryId  = "category_id"
)

// PromotionProgramType is how a promotion is triggered and computed (BR §3.1).
type PromotionProgramType string

const (
	PromotionTypeAutomatic       = PromotionProgramType("automatic")
	PromotionTypeDiscountCode    = PromotionProgramType("discount_code")
	PromotionTypeCoupon          = PromotionProgramType("coupon")
	PromotionTypeBuyXGetY        = PromotionProgramType("buy_x_get_y")
	PromotionTypeNextOrderCoupon = PromotionProgramType("next_order_coupon")
)

// TakesVouchers reports whether a promotion of this type is unlocked by a voucher (BR-V01).
func (this PromotionProgramType) TakesVouchers() bool {
	return this == PromotionTypeDiscountCode || this == PromotionTypeCoupon ||
		this == PromotionTypeNextOrderCoupon
}

// PromotionStatus is the business lifecycle (BR-L01); expired and exhausted are derived.
type PromotionStatus string

const (
	PromotionStatusDraft     = PromotionStatus("draft")
	PromotionStatusActive    = PromotionStatus("active")
	PromotionStatusSuspended = PromotionStatus("suspended")
)

// PromotionRulePhase says which order a rule qualifies (BR-V08).
type PromotionRulePhase string

const (
	PromotionRulePhaseApply = PromotionRulePhase("apply")
	PromotionRulePhaseIssue = PromotionRulePhase("issue")
)

// PromotionRewardType is what the customer receives (BR §5).
type PromotionRewardType string

const (
	PromotionRewardDiscount     = PromotionRewardType("discount")
	PromotionRewardFreeProduct  = PromotionRewardType("free_product")
	PromotionRewardFreeShipping = PromotionRewardType("free_shipping")
)

const (
	PromotionDiscountScopeOrder            = "order"
	PromotionDiscountScopeSpecificProducts = "specific_products"
	PromotionDiscountScopeCheapestProduct  = "cheapest_product"
)

//go:embed sales_promotion.json
var salesPromotionSchemaJson string

func SalesPromotionSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesPromotionSchemaJson)
}

//go:embed sales_promotion_channel.json
var salesPromotionChannelSchemaJson string

func SalesPromotionChannelSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesPromotionChannelSchemaJson)
}

//go:embed sales_promotion_point.json
var salesPromotionPointSchemaJson string

func SalesPromotionPointSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesPromotionPointSchemaJson)
}

//go:embed sales_promotion_pricelist.json
var salesPromotionPricelistSchemaJson string

func SalesPromotionPricelistSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesPromotionPricelistSchemaJson)
}

//go:embed sales_promotion_rule.json
var salesPromotionRuleSchemaJson string

func SalesPromotionRuleSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesPromotionRuleSchemaJson)
}

//go:embed sales_promotion_rule_product.json
var salesPromotionRuleProductSchemaJson string

func SalesPromotionRuleProductSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesPromotionRuleProductSchemaJson)
}

//go:embed sales_promotion_rule_category.json
var salesPromotionRuleCategorySchemaJson string

func SalesPromotionRuleCategorySchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesPromotionRuleCategorySchemaJson)
}

//go:embed sales_promotion_reward.json
var salesPromotionRewardSchemaJson string

func SalesPromotionRewardSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesPromotionRewardSchemaJson)
}

//go:embed sales_promotion_reward_product.json
var salesPromotionRewardProductSchemaJson string

func SalesPromotionRewardProductSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesPromotionRewardProductSchemaJson)
}

//go:embed sales_promotion_reward_category.json
var salesPromotionRewardCategorySchemaJson string

func SalesPromotionRewardCategorySchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesPromotionRewardCategorySchemaJson)
}

// SalesPromotion is one promotion's configuration, not a use of it.
type SalesPromotion struct {
	basemodel.DynamicModelBase
}

func NewSalesPromotionFrom(src dmodel.DynamicFields) *SalesPromotion {
	return &SalesPromotion{basemodel.NewDynamicModel(src)}
}

func (this SalesPromotion) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this SalesPromotion) GetProgramType() *string {
	return this.GetFieldData().GetString(SalesPromotionFieldProgramType)
}

func (this SalesPromotion) GetStatus() *string {
	return this.GetFieldData().GetString(SalesPromotionFieldStatus)
}

// SalesPromotionReward is one benefit a promotion offers.
type SalesPromotionReward struct {
	basemodel.DynamicModelBase
}

func NewSalesPromotionRewardFrom(src dmodel.DynamicFields) *SalesPromotionReward {
	return &SalesPromotionReward{basemodel.NewDynamicModel(src)}
}

func (this SalesPromotionReward) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

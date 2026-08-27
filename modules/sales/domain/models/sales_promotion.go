package models

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// The promotion schemas. One engine serves automatic promotions, conditional bundle pricing (BR §20)
// and vouchers (BR §25) — BR §21–22 forbid three separate engines, and the difference between them is
// data rather than code: an automatic program and a voucher program differ only in activation_type.

const (
	SalesPromotionProgramSchemaName = "sales_promotion_program"

	SalesPromotionProgramFieldId                    = "id"
	SalesPromotionProgramFieldOrgId                 = "org_id"
	SalesPromotionProgramFieldCode                  = "code"
	SalesPromotionProgramFieldName                  = "name"
	SalesPromotionProgramFieldActivationType        = "activation_type"
	SalesPromotionProgramFieldPriority              = "priority"
	SalesPromotionProgramFieldValidFrom             = "valid_from"
	SalesPromotionProgramFieldValidUntil            = "valid_until"
	SalesPromotionProgramFieldStackPolicy           = "stack_policy"
	SalesPromotionProgramFieldExclusiveGroup        = "exclusive_group"
	SalesPromotionProgramFieldUsageLimit            = "usage_limit"
	SalesPromotionProgramFieldUsageLimitPerCustomer = "usage_limit_per_customer"
	SalesPromotionProgramFieldReturnBehavior        = "return_behavior"
	SalesPromotionProgramFieldRestoreOnFullReturn   = "restore_on_full_return"
	SalesPromotionProgramFieldRestoreOnPartialRet   = "restore_on_partial_return"
)

const (
	SalesPromotionConditionGroupSchemaName = "sales_promotion_condition_group"

	SalesPromotionConditionGroupFieldId        = "id"
	SalesPromotionConditionGroupFieldProgramId = "sales_promotion_program_id"
	SalesPromotionConditionGroupFieldSequence  = "sequence"
)

const (
	SalesPromotionConditionSchemaName = "sales_promotion_condition"

	SalesPromotionConditionFieldId            = "id"
	SalesPromotionConditionFieldGroupId       = "group_id"
	SalesPromotionConditionFieldConditionType = "condition_type"
	SalesPromotionConditionFieldOperator      = "operator"
	SalesPromotionConditionFieldValueText     = "value_text"
	SalesPromotionConditionFieldValueDecimal  = "value_decimal"
	SalesPromotionConditionFieldValueFrom     = "value_from"
	SalesPromotionConditionFieldValueTo       = "value_to"
)

const (
	SalesPromotionConditionTargetSchemaName = "sales_promotion_condition_target"

	SalesPromotionConditionTargetFieldId          = "id"
	SalesPromotionConditionTargetFieldConditionId = "condition_id"
	SalesPromotionConditionTargetFieldTargetType  = "target_type"
	SalesPromotionConditionTargetFieldTargetId    = "target_id"
)

const (
	SalesPromotionRewardSchemaName = "sales_promotion_reward"

	SalesPromotionRewardFieldId          = "id"
	SalesPromotionRewardFieldProgramId   = "sales_promotion_program_id"
	SalesPromotionRewardFieldSequence    = "sequence"
	SalesPromotionRewardFieldRewardType  = "reward_type"
	SalesPromotionRewardFieldValue       = "value"
	SalesPromotionRewardFieldTargetScope = "target_scope"
)

const (
	SalesPromotionCompatibilitySchemaName = "sales_promotion_compatibility"

	SalesPromotionCompatibilityFieldId            = "id"
	SalesPromotionCompatibilityFieldProgramAId    = "program_a_id"
	SalesPromotionCompatibilityFieldProgramBId    = "program_b_id"
	SalesPromotionCompatibilityFieldCompatibility = "compatibility"
)

// PromotionActivationType is how a program comes into play.
type PromotionActivationType string

const (
	// PromotionActivationAutomatic applies whenever its conditions hold.
	PromotionActivationAutomatic = PromotionActivationType("automatic")
	// PromotionActivationVoucherCode applies only when the customer presents a code.
	PromotionActivationVoucherCode = PromotionActivationType("voucher_code")
)

// PromotionStackPolicy is how a program combines with others when no explicit compatibility row
// decides (D-08). It is the FALLBACK in the D-09 resolution order, never the first word.
type PromotionStackPolicy string

const (
	PromotionStackStackable = PromotionStackPolicy("stackable")
	// PromotionStackExclusive applies alone on the order.
	PromotionStackExclusive = PromotionStackPolicy("exclusive")
	// PromotionStackExclusiveWithinGroup excludes only programs sharing its exclusive_group.
	PromotionStackExclusiveWithinGroup = PromotionStackPolicy("exclusive_within_group")
)

// PromotionCompatibility is an explicit pairwise directive (D-09).
type PromotionCompatibility string

const (
	PromotionCompatibilityAllowed = PromotionCompatibility("allowed")
	// PromotionCompatibilityDenied wins over everything, always.
	PromotionCompatibilityDenied = PromotionCompatibility("denied")
)

// PromotionRewardType is what the customer gets (BR §24).
type PromotionRewardType string

const (
	PromotionRewardPercentageDiscount  = PromotionRewardType("percentage_discount")
	PromotionRewardFixedAmountDiscount = PromotionRewardType("fixed_amount_discount")
	PromotionRewardFixedProductPrice   = PromotionRewardType("fixed_product_price")
	// PromotionRewardFreeQuantity becomes a separate order line at zero price (D-11), because
	// Inventory must physically fulfil the free item.
	PromotionRewardFreeQuantity = PromotionRewardType("free_quantity")
)

// PromotionReturnBehavior is what happens to a discount when part of the order comes back (BR §31).
type PromotionReturnBehavior string

const (
	PromotionReturnPreserveOriginal = PromotionReturnBehavior("preserve_original_discount")
	// PromotionReturnRevalidate re-runs the engine over the remaining basket and reclaims the
	// difference as a clawback adjustment (D-12).
	PromotionReturnRevalidate = PromotionReturnBehavior("revalidate_and_clawback")
)

// PromotionConditionType is what a condition tests (BR §23).
//
// Customer and customer group are deliberately absent: BR §23 marks them future, and an enum value
// nothing evaluates reads as a supported feature that silently never matches.
type PromotionConditionType string

const (
	PromotionConditionProductVariant  = PromotionConditionType("product_variant")
	PromotionConditionProductCategory = PromotionConditionType("product_category")
	PromotionConditionQuantity        = PromotionConditionType("quantity")
	PromotionConditionOrderSubtotal   = PromotionConditionType("order_subtotal")
	PromotionConditionTotalQuantity   = PromotionConditionType("total_quantity")
	PromotionConditionValidFrom       = PromotionConditionType("valid_from")
	PromotionConditionValidUntil      = PromotionConditionType("valid_until")
	PromotionConditionDayOfWeek       = PromotionConditionType("day_of_week")
	PromotionConditionTimeOfDay       = PromotionConditionType("time_of_day")
	PromotionConditionSalesChannel    = PromotionConditionType("sales_channel")
	PromotionConditionSalesPoint      = PromotionConditionType("sales_point")
)

// PromotionOperator is how a condition compares (D-07).
type PromotionOperator string

const (
	PromotionOperatorEq      = PromotionOperator("eq")
	PromotionOperatorNe      = PromotionOperator("ne")
	PromotionOperatorGt      = PromotionOperator("gt")
	PromotionOperatorGte     = PromotionOperator("gte")
	PromotionOperatorLt      = PromotionOperator("lt")
	PromotionOperatorLte     = PromotionOperator("lte")
	PromotionOperatorIn      = PromotionOperator("in")
	PromotionOperatorNotIn   = PromotionOperator("not_in")
	PromotionOperatorBetween = PromotionOperator("between")
)

//go:embed sales_promotion_program.json
var salesPromotionProgramSchemaJson string

func SalesPromotionProgramSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesPromotionProgramSchemaJson)
}

//go:embed sales_promotion_condition_group.json
var salesPromotionConditionGroupSchemaJson string

func SalesPromotionConditionGroupSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesPromotionConditionGroupSchemaJson)
}

//go:embed sales_promotion_condition.json
var salesPromotionConditionSchemaJson string

func SalesPromotionConditionSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesPromotionConditionSchemaJson)
}

//go:embed sales_promotion_condition_target.json
var salesPromotionConditionTargetSchemaJson string

func SalesPromotionConditionTargetSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesPromotionConditionTargetSchemaJson)
}

//go:embed sales_promotion_reward.json
var salesPromotionRewardSchemaJson string

func SalesPromotionRewardSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesPromotionRewardSchemaJson)
}

//go:embed sales_promotion_compatibility.json
var salesPromotionCompatibilitySchemaJson string

func SalesPromotionCompatibilitySchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesPromotionCompatibilitySchemaJson)
}

// SalesPromotionProgram is one campaign: what has to be true, and what the customer then gets.
type SalesPromotionProgram struct {
	basemodel.DynamicModelBase
}

func NewSalesPromotionProgram() *SalesPromotionProgram {
	return &SalesPromotionProgram{basemodel.NewDynamicModel()}
}

func NewSalesPromotionProgramFrom(src dmodel.DynamicFields) *SalesPromotionProgram {
	return &SalesPromotionProgram{basemodel.NewDynamicModel(src)}
}

func (this SalesPromotionProgram) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this SalesPromotionProgram) GetCode() *string {
	return this.GetFieldData().GetString(SalesPromotionProgramFieldCode)
}

func (this SalesPromotionProgram) GetName() *string {
	return this.GetFieldData().GetString(SalesPromotionProgramFieldName)
}

func (this SalesPromotionProgram) GetActivationType() *string {
	return this.GetFieldData().GetString(SalesPromotionProgramFieldActivationType)
}

func (this SalesPromotionProgram) GetPriority() *int32 {
	return this.GetFieldData().GetInt32(SalesPromotionProgramFieldPriority)
}

func (this SalesPromotionProgram) GetValidFrom() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(SalesPromotionProgramFieldValidFrom)
}

func (this SalesPromotionProgram) GetValidUntil() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(SalesPromotionProgramFieldValidUntil)
}

func (this SalesPromotionProgram) GetStackPolicy() *string {
	return this.GetFieldData().GetString(SalesPromotionProgramFieldStackPolicy)
}

func (this SalesPromotionProgram) GetExclusiveGroup() *string {
	return this.GetFieldData().GetString(SalesPromotionProgramFieldExclusiveGroup)
}

func (this SalesPromotionProgram) GetUsageLimit() *int32 {
	return this.GetFieldData().GetInt32(SalesPromotionProgramFieldUsageLimit)
}

func (this SalesPromotionProgram) GetReturnBehavior() *string {
	return this.GetFieldData().GetString(SalesPromotionProgramFieldReturnBehavior)
}

func (this SalesPromotionProgram) GetIsArchived() *bool {
	return this.GetFieldData().GetBool(basemodel.FieldIsArchived)
}

// RequiresVoucherCode reports whether this program only applies when a code is presented.
func (this SalesPromotionProgram) RequiresVoucherCode() bool {
	activation := this.GetActivationType()
	return activation != nil &&
		PromotionActivationType(*activation) == PromotionActivationVoucherCode
}

// SalesPromotionReward is one thing a program gives.
type SalesPromotionReward struct {
	basemodel.DynamicModelBase
}

func NewSalesPromotionRewardFrom(src dmodel.DynamicFields) *SalesPromotionReward {
	return &SalesPromotionReward{basemodel.NewDynamicModel(src)}
}

func (this SalesPromotionReward) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this SalesPromotionReward) GetProgramId() *model.Id {
	return this.GetFieldData().GetModelId(SalesPromotionRewardFieldProgramId)
}

func (this SalesPromotionReward) GetSequence() *int32 {
	return this.GetFieldData().GetInt32(SalesPromotionRewardFieldSequence)
}

func (this SalesPromotionReward) GetRewardType() *string {
	return this.GetFieldData().GetString(SalesPromotionRewardFieldRewardType)
}

func (this SalesPromotionReward) GetValue() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesPromotionRewardFieldValue)
}

func (this SalesPromotionReward) GetTargetScope() *string {
	return this.GetFieldData().GetString(SalesPromotionRewardFieldTargetScope)
}

// SalesPromotionCondition is one test a program applies.
type SalesPromotionCondition struct {
	basemodel.DynamicModelBase
}

func NewSalesPromotionConditionFrom(src dmodel.DynamicFields) *SalesPromotionCondition {
	return &SalesPromotionCondition{basemodel.NewDynamicModel(src)}
}

func (this SalesPromotionCondition) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this SalesPromotionCondition) GetGroupId() *model.Id {
	return this.GetFieldData().GetModelId(SalesPromotionConditionFieldGroupId)
}

func (this SalesPromotionCondition) GetConditionType() *string {
	return this.GetFieldData().GetString(SalesPromotionConditionFieldConditionType)
}

func (this SalesPromotionCondition) GetOperator() *string {
	return this.GetFieldData().GetString(SalesPromotionConditionFieldOperator)
}

func (this SalesPromotionCondition) GetValueText() *string {
	return this.GetFieldData().GetString(SalesPromotionConditionFieldValueText)
}

func (this SalesPromotionCondition) GetValueDecimal() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesPromotionConditionFieldValueDecimal)
}

func (this SalesPromotionCondition) GetValueFrom() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesPromotionConditionFieldValueFrom)
}

func (this SalesPromotionCondition) GetValueTo() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesPromotionConditionFieldValueTo)
}

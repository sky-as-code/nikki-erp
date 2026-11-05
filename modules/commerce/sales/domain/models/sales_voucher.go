package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// The voucher and the order-side promotion records (BR §8.7-§8.10). A voucher only unlocks its
// promotion; it carries no value of its own. A draft order holds attached vouchers and reward
// choices; applications, allocations and issuances exist only from Confirm onward.

const (
	SalesVoucherSchemaName = "sales_voucher"

	SalesVoucherFieldId               = basemodel.FieldId
	SalesVoucherFieldOrgId            = basemodel.FieldOrgId
	SalesVoucherFieldPromotionId      = "promotion_id"
	SalesVoucherFieldCode             = "code"
	SalesVoucherFieldNormalizedCode   = "normalized_code"
	SalesVoucherFieldVoucherType      = "voucher_type"
	SalesVoucherFieldStatus           = "status"
	SalesVoucherFieldCustomerId       = "customer_id"
	SalesVoucherFieldValidFrom        = "valid_from"
	SalesVoucherFieldExpiresAt        = "expires_at"
	SalesVoucherFieldUsageLimit       = "usage_limit"
	SalesVoucherFieldUsedCount        = "used_count"
	SalesVoucherFieldSourceOrderId    = "source_order_id"
	SalesVoucherFieldIssuedAt         = "issued_at"
	SalesVoucherFieldRevokedAt        = "revoked_at"
	SalesVoucherFieldRevokedBy        = "revoked_by"
	SalesVoucherFieldRevocationReason = "revocation_reason"
	SalesVoucherFieldRevocationNote   = "revocation_note"
)

const (
	SalesOrderVoucherSchemaName = "sales_order_voucher"

	SalesOrderVoucherFieldOrderId     = "order_id"
	SalesOrderVoucherFieldVoucherId   = "voucher_id"
	SalesOrderVoucherFieldPromotionId = "promotion_id"
	SalesOrderVoucherFieldAttachedAt  = "attached_at"
	SalesOrderVoucherFieldAttachedBy  = "attached_by"
)

const (
	SalesOrderPromotionSelectionSchemaName = "sales_order_promotion_selection"

	SalesOrderPromotionSelectionFieldOrderId     = "order_id"
	SalesOrderPromotionSelectionFieldPromotionId = "promotion_id"
	SalesOrderPromotionSelectionFieldRewardId    = "reward_id"
	SalesOrderPromotionSelectionFieldDecision    = "decision"
)

const (
	SalesPromotionApplicationSchemaName = "sales_promotion_application"

	SalesPromotionApplicationFieldOrderId               = "order_id"
	SalesPromotionApplicationFieldPromotionId           = "promotion_id"
	SalesPromotionApplicationFieldRewardId              = "reward_id"
	SalesPromotionApplicationFieldVoucherId             = "voucher_id"
	SalesPromotionApplicationFieldStatus                = "status"
	SalesPromotionApplicationFieldIsConfirmed           = "is_confirmed"
	SalesPromotionApplicationFieldConfirmationRevision  = "confirmation_revision"
	SalesPromotionApplicationFieldConfigVersion         = "config_version"
	SalesPromotionApplicationFieldQualificationSnapshot = "qualification_snapshot"
	SalesPromotionApplicationFieldRewardSnapshot        = "reward_snapshot"
	SalesPromotionApplicationFieldDiscountUntaxed       = "discount_untaxed_amount"
	SalesPromotionApplicationFieldTaxReduction          = "tax_reduction_amount"
	SalesPromotionApplicationFieldDiscountTotal         = "discount_total_amount"
	SalesPromotionApplicationFieldAppliedAt             = "applied_at"
	SalesPromotionApplicationFieldReversedAt            = "reversed_at"
	SalesPromotionApplicationFieldReversalReason        = "reversal_reason"
)

const (
	SalesPromotionAllocationSchemaName = "sales_promotion_allocation"

	SalesPromotionAllocationFieldApplicationId       = "application_id"
	SalesPromotionAllocationFieldOrderLineId         = "order_line_id"
	SalesPromotionAllocationFieldAllocationType      = "allocation_type"
	SalesPromotionAllocationFieldEligibleQuantity    = "eligible_quantity"
	SalesPromotionAllocationFieldUomId               = "uom_id"
	SalesPromotionAllocationFieldBaseUntaxed         = "base_untaxed_amount"
	SalesPromotionAllocationFieldDiscountUntaxed     = "discount_untaxed_amount"
	SalesPromotionAllocationFieldTaxReduction        = "tax_reduction_amount"
	SalesPromotionAllocationFieldDiscountTotal       = "discount_total_amount"
	SalesPromotionAllocationFieldCalculationSequence = "calculation_sequence"
)

const (
	SalesVoucherIssuanceSchemaName = "sales_voucher_issuance"

	SalesVoucherIssuanceFieldSourceOrderId     = "source_order_id"
	SalesVoucherIssuanceFieldPromotionId       = "promotion_id"
	SalesVoucherIssuanceFieldVoucherId         = "voucher_id"
	SalesVoucherIssuanceFieldIssueRuleSnapshot = "issue_rule_snapshot"
	SalesVoucherIssuanceFieldIssuedAt          = "issued_at"
)

// VoucherType is derived from the promotion type when the voucher is created (BR-V02).
type VoucherType string

const (
	VoucherTypeReusable  = VoucherType("reusable")
	VoucherTypeSingleUse = VoucherType("single_use")
)

// VoucherStatus is stored; not started, expired and exhausted are derived (BR-V05).
type VoucherStatus string

const (
	VoucherStatusActive  = VoucherStatus("active")
	VoucherStatusRevoked = VoucherStatus("revoked")
)

const (
	VoucherRevocationManual               = "manual"
	VoucherRevocationSourceOrderCancelled = "source_order_cancelled"
)

// PromotionApplicationStatus is the lifecycle of a confirmed benefit (BR-T07).
type PromotionApplicationStatus string

const (
	PromotionApplicationConfirmed = PromotionApplicationStatus("confirmed")
	PromotionApplicationReversed  = PromotionApplicationStatus("reversed")
)

//go:embed sales_voucher.json
var salesVoucherSchemaJson string

func SalesVoucherSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesVoucherSchemaJson)
}

//go:embed sales_order_voucher.json
var salesOrderVoucherSchemaJson string

func SalesOrderVoucherSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesOrderVoucherSchemaJson)
}

//go:embed sales_order_promotion_selection.json
var salesOrderPromotionSelectionSchemaJson string

func SalesOrderPromotionSelectionSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesOrderPromotionSelectionSchemaJson)
}

//go:embed sales_promotion_application.json
var salesPromotionApplicationSchemaJson string

func SalesPromotionApplicationSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesPromotionApplicationSchemaJson)
}

//go:embed sales_promotion_allocation.json
var salesPromotionAllocationSchemaJson string

func SalesPromotionAllocationSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesPromotionAllocationSchemaJson)
}

//go:embed sales_voucher_issuance.json
var salesVoucherIssuanceSchemaJson string

func SalesVoucherIssuanceSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesVoucherIssuanceSchemaJson)
}

// SalesVoucher is one code that unlocks a promotion.
type SalesVoucher struct {
	basemodel.DynamicModelBase
}

func NewSalesVoucherFrom(src dmodel.DynamicFields) *SalesVoucher {
	return &SalesVoucher{basemodel.NewDynamicModel(src)}
}

func (this SalesVoucher) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this SalesVoucher) GetPromotionId() *model.Id {
	return this.GetFieldData().GetModelId(SalesVoucherFieldPromotionId)
}

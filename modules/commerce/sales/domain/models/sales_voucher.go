package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// The voucher pair: a code, and the ledger of its uses. A usage counter alone is insufficient —
// it cannot be audited, and cannot say which orders consumed a voucher.

const (
	SalesVoucherCodeSchemaName = "sales_voucher_code"

	SalesVoucherCodeFieldId         = basemodel.FieldId
	SalesVoucherCodeFieldOrgId      = basemodel.FieldOrgId
	SalesVoucherCodeFieldCode       = "code"
	SalesVoucherCodeFieldProgramId  = "sales_promotion_program_id"
	SalesVoucherCodeFieldValidFrom  = "valid_from"
	SalesVoucherCodeFieldValidUntil = "valid_until"
	SalesVoucherCodeFieldUsageLimit = "usage_limit"
	SalesVoucherCodeFieldUsageCount = "usage_count"
	SalesVoucherCodeFieldStatus     = "status"
	SalesVoucherCodeFieldIsArchived = basemodel.FieldIsArchived

	SalesVoucherCodeEdgeProgram = "sales_promotion_program"
)

const (
	SalesVoucherRedemptionSchemaName = "sales_voucher_redemption"

	SalesVoucherRedemptionFieldId            = basemodel.FieldId
	SalesVoucherRedemptionFieldOrgId         = basemodel.FieldOrgId
	SalesVoucherRedemptionFieldVoucherCodeId = "voucher_code_id"
	SalesVoucherRedemptionFieldSalesOrderId  = "sales_order_id"
	SalesVoucherRedemptionFieldStatus        = "status"
	SalesVoucherRedemptionFieldReservedAt    = "reserved_at"
	SalesVoucherRedemptionFieldRedeemedAt    = "redeemed_at"
	SalesVoucherRedemptionFieldReleasedAt    = "released_at"
	SalesVoucherRedemptionFieldReversedAt    = "reversed_at"

	SalesVoucherRedemptionEdgeVoucherCode = "voucher_code"
	SalesVoucherRedemptionEdgeSalesOrder  = "sales_order"
)

//go:embed sales_voucher_code.json
var salesVoucherCodeSchemaJson string

func SalesVoucherCodeSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesVoucherCodeSchemaJson)
}

//go:embed sales_voucher_redemption.json
var salesVoucherRedemptionSchemaJson string

func SalesVoucherRedemptionSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesVoucherRedemptionSchemaJson)
}

// SalesVoucherCode is a credential pointing at a promotion program. It carries no rules of its own:
// every condition, reward, compatibility rule and priority lives on the program.
type SalesVoucherCode struct {
	basemodel.DynamicModelBase
}

func NewSalesVoucherCode() *SalesVoucherCode {
	return &SalesVoucherCode{basemodel.NewDynamicModel()}
}

func NewSalesVoucherCodeFrom(src dmodel.DynamicFields) *SalesVoucherCode {
	return &SalesVoucherCode{basemodel.NewDynamicModel(src)}
}

func (this SalesVoucherCode) GetId() *model.Id {
	return this.GetFieldData().GetModelId(SalesVoucherCodeFieldId)
}

func (this SalesVoucherCode) GetCode() *string {
	return this.GetFieldData().GetString(SalesVoucherCodeFieldCode)
}

func (this *SalesVoucherCode) SetCode(code *string) {
	this.GetFieldData().SetString(SalesVoucherCodeFieldCode, code)
}

func (this SalesVoucherCode) GetProgramId() *model.Id {
	return this.GetFieldData().GetModelId(SalesVoucherCodeFieldProgramId)
}

func (this *SalesVoucherCode) SetProgramId(id *model.Id) {
	this.GetFieldData().SetModelId(SalesVoucherCodeFieldProgramId, id)
}

func (this SalesVoucherCode) GetValidFrom() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(SalesVoucherCodeFieldValidFrom)
}

func (this *SalesVoucherCode) SetValidFrom(at *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(SalesVoucherCodeFieldValidFrom, at)
}

func (this SalesVoucherCode) GetValidUntil() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(SalesVoucherCodeFieldValidUntil)
}

func (this *SalesVoucherCode) SetValidUntil(at *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(SalesVoucherCodeFieldValidUntil, at)
}

func (this SalesVoucherCode) GetUsageLimit() *int32 {
	return this.GetFieldData().GetInt32(SalesVoucherCodeFieldUsageLimit)
}

func (this *SalesVoucherCode) SetUsageLimit(limit *int32) {
	this.GetFieldData().SetInt32(SalesVoucherCodeFieldUsageLimit, limit)
}

func (this SalesVoucherCode) GetUsageCount() *int32 {
	return this.GetFieldData().GetInt32(SalesVoucherCodeFieldUsageCount)
}

func (this *SalesVoucherCode) SetUsageCount(count *int32) {
	this.GetFieldData().SetInt32(SalesVoucherCodeFieldUsageCount, count)
}

func (this SalesVoucherCode) GetStatus() *string {
	return this.GetFieldData().GetString(SalesVoucherCodeFieldStatus)
}

func (this *SalesVoucherCode) SetStatus(status *string) {
	this.GetFieldData().SetString(SalesVoucherCodeFieldStatus, status)
}

func (this SalesVoucherCode) GetIsArchived() *bool {
	return this.GetFieldData().GetBool(SalesVoucherCodeFieldIsArchived)
}

// HasUsesRemaining reports whether the code has any use left. A null usage_limit means unlimited,
// so this is not a plain comparison. It is only the read of the invariant, for explaining a refusal;
// enforcement is the conditional UPDATE in the redemption path, since a check before a separate
// write can be overtaken.
func (this SalesVoucherCode) HasUsesRemaining() bool {
	limit := this.GetUsageLimit()
	if limit == nil {
		return true
	}
	count := this.GetUsageCount()
	if count == nil {
		return *limit > 0
	}
	return *count < *limit
}

// IsUsableAt reports whether the code may be applied at the given instant: status, archival,
// validity window and usage count only. Program eligibility, channel rules and compatibility with
// other vouchers need records this model does not hold, so they stay with the apply operation. The
// window is half-open - valid_from inclusive, valid_until exclusive.
func (this SalesVoucherCode) IsUsableAt(unixSeconds int64) bool {
	if archived := this.GetIsArchived(); archived != nil && *archived {
		return false
	}
	if status := this.GetStatus(); status == nil ||
		*status != string(VoucherCodeStatusActive) {
		return false
	}
	if from := this.GetValidFrom(); from != nil && unixSeconds < from.GoTime().Unix() {
		return false
	}
	if until := this.GetValidUntil(); until != nil && unixSeconds >= until.GoTime().Unix() {
		return false
	}
	return this.HasUsesRemaining()
}

// SalesVoucherRedemption is one code's use on one order. The table exists because a counter cannot
// say which orders consumed a voucher, nor hold a use while an order is still a draft.
type SalesVoucherRedemption struct {
	basemodel.DynamicModelBase
}

func NewSalesVoucherRedemption() *SalesVoucherRedemption {
	return &SalesVoucherRedemption{basemodel.NewDynamicModel()}
}

func NewSalesVoucherRedemptionFrom(src dmodel.DynamicFields) *SalesVoucherRedemption {
	return &SalesVoucherRedemption{basemodel.NewDynamicModel(src)}
}

func (this SalesVoucherRedemption) GetId() *model.Id {
	return this.GetFieldData().GetModelId(SalesVoucherRedemptionFieldId)
}

func (this SalesVoucherRedemption) GetVoucherCodeId() *model.Id {
	return this.GetFieldData().GetModelId(SalesVoucherRedemptionFieldVoucherCodeId)
}

func (this *SalesVoucherRedemption) SetVoucherCodeId(id *model.Id) {
	this.GetFieldData().SetModelId(SalesVoucherRedemptionFieldVoucherCodeId, id)
}

func (this SalesVoucherRedemption) GetSalesOrderId() *model.Id {
	return this.GetFieldData().GetModelId(SalesVoucherRedemptionFieldSalesOrderId)
}

func (this *SalesVoucherRedemption) SetSalesOrderId(id *model.Id) {
	this.GetFieldData().SetModelId(SalesVoucherRedemptionFieldSalesOrderId, id)
}

func (this SalesVoucherRedemption) GetStatus() *string {
	return this.GetFieldData().GetString(SalesVoucherRedemptionFieldStatus)
}

func (this *SalesVoucherRedemption) SetStatus(status *string) {
	this.GetFieldData().SetString(SalesVoucherRedemptionFieldStatus, status)
}

func (this SalesVoucherRedemption) GetReservedAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(SalesVoucherRedemptionFieldReservedAt)
}

func (this *SalesVoucherRedemption) SetReservedAt(at *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(SalesVoucherRedemptionFieldReservedAt, at)
}

func (this SalesVoucherRedemption) GetRedeemedAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(SalesVoucherRedemptionFieldRedeemedAt)
}

func (this *SalesVoucherRedemption) SetRedeemedAt(at *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(SalesVoucherRedemptionFieldRedeemedAt, at)
}

func (this SalesVoucherRedemption) GetReleasedAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(SalesVoucherRedemptionFieldReleasedAt)
}

func (this *SalesVoucherRedemption) SetReleasedAt(at *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(SalesVoucherRedemptionFieldReleasedAt, at)
}

func (this SalesVoucherRedemption) GetReversedAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(SalesVoucherRedemptionFieldReversedAt)
}

func (this *SalesVoucherRedemption) SetReversedAt(at *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(SalesVoucherRedemptionFieldReversedAt, at)
}

// HoldsAUse reports whether this redemption is currently consuming one of the code's uses. A
// reservation counts: a code in someone's draft basket is not available to the next customer.
func (this SalesVoucherRedemption) HoldsAUse() bool {
	status := this.GetStatus()
	if status == nil {
		return false
	}
	return *status == string(VoucherRedemptionStatusReserved) ||
		*status == string(VoucherRedemptionStatusRedeemed)
}

package models

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	SalesFulfillmentAttemptItemSchemaName = "sales_fulfillment_attempt_item"

	SalesFulfillmentAttemptItemFieldId                = "id"
	SalesFulfillmentAttemptItemFieldOrgId             = "org_id"
	SalesFulfillmentAttemptItemFieldAttemptId         = "attempt_id"
	SalesFulfillmentAttemptItemFieldFulfillmentItemId = "fulfillment_item_id"
	SalesFulfillmentAttemptItemFieldAttemptedQty      = "attempted_qty"
	SalesFulfillmentAttemptItemFieldDispensedQty      = "dispensed_qty"
	SalesFulfillmentAttemptItemFieldFailedQty         = "failed_qty"
	SalesFulfillmentAttemptItemFieldItemResult        = "item_result"
	SalesFulfillmentAttemptItemFieldFailureCode       = "failure_code"
	SalesFulfillmentAttemptItemFieldFailureMessage    = "failure_message"
)

//go:embed sales_fulfillment_attempt_item.json
var salesFulfillmentAttemptItemSchemaJson string

func SalesFulfillmentAttemptItemSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesFulfillmentAttemptItemSchemaJson)
}

type SalesFulfillmentAttemptItem struct {
	basemodel.DynamicModelBase
}

func NewSalesFulfillmentAttemptItem() *SalesFulfillmentAttemptItem {
	return &SalesFulfillmentAttemptItem{basemodel.NewDynamicModel()}
}

func NewSalesFulfillmentAttemptItemFrom(src dmodel.DynamicFields) *SalesFulfillmentAttemptItem {
	return &SalesFulfillmentAttemptItem{basemodel.NewDynamicModel(src)}
}

func (this SalesFulfillmentAttemptItem) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this *SalesFulfillmentAttemptItem) SetId(id *model.Id) {
	this.GetFieldData().SetModelId(basemodel.FieldId, id)
}

func (this SalesFulfillmentAttemptItem) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(SalesFulfillmentAttemptItemFieldOrgId)
}

func (this *SalesFulfillmentAttemptItem) SetOrgId(value *model.Id) {
	this.GetFieldData().SetModelId(SalesFulfillmentAttemptItemFieldOrgId, value)
}

func (this SalesFulfillmentAttemptItem) GetAttemptId() *model.Id {
	return this.GetFieldData().GetModelId(SalesFulfillmentAttemptItemFieldAttemptId)
}

func (this *SalesFulfillmentAttemptItem) SetAttemptId(value *model.Id) {
	this.GetFieldData().SetModelId(SalesFulfillmentAttemptItemFieldAttemptId, value)
}

func (this SalesFulfillmentAttemptItem) GetFulfillmentItemId() *model.Id {
	return this.GetFieldData().GetModelId(SalesFulfillmentAttemptItemFieldFulfillmentItemId)
}

func (this *SalesFulfillmentAttemptItem) SetFulfillmentItemId(value *model.Id) {
	this.GetFieldData().SetModelId(SalesFulfillmentAttemptItemFieldFulfillmentItemId, value)
}

func (this SalesFulfillmentAttemptItem) GetAttemptedQty() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesFulfillmentAttemptItemFieldAttemptedQty)
}

func (this *SalesFulfillmentAttemptItem) SetAttemptedQty(value *decimal.Decimal) {
	this.GetFieldData().SetDecimal(SalesFulfillmentAttemptItemFieldAttemptedQty, value)
}

func (this SalesFulfillmentAttemptItem) GetDispensedQty() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesFulfillmentAttemptItemFieldDispensedQty)
}

func (this *SalesFulfillmentAttemptItem) SetDispensedQty(value *decimal.Decimal) {
	this.GetFieldData().SetDecimal(SalesFulfillmentAttemptItemFieldDispensedQty, value)
}

func (this SalesFulfillmentAttemptItem) GetFailedQty() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesFulfillmentAttemptItemFieldFailedQty)
}

func (this *SalesFulfillmentAttemptItem) SetFailedQty(value *decimal.Decimal) {
	this.GetFieldData().SetDecimal(SalesFulfillmentAttemptItemFieldFailedQty, value)
}

func (this SalesFulfillmentAttemptItem) GetItemResult() *string {
	return this.GetFieldData().GetString(SalesFulfillmentAttemptItemFieldItemResult)
}

func (this *SalesFulfillmentAttemptItem) SetItemResult(value *string) {
	this.GetFieldData().SetString(SalesFulfillmentAttemptItemFieldItemResult, value)
}

func (this SalesFulfillmentAttemptItem) GetFailureCode() *string {
	return this.GetFieldData().GetString(SalesFulfillmentAttemptItemFieldFailureCode)
}

func (this *SalesFulfillmentAttemptItem) SetFailureCode(value *string) {
	this.GetFieldData().SetString(SalesFulfillmentAttemptItemFieldFailureCode, value)
}

func (this SalesFulfillmentAttemptItem) GetFailureMessage() *string {
	return this.GetFieldData().GetString(SalesFulfillmentAttemptItemFieldFailureMessage)
}

func (this *SalesFulfillmentAttemptItem) SetFailureMessage(value *string) {
	this.GetFieldData().SetString(SalesFulfillmentAttemptItemFieldFailureMessage, value)
}

// DeriveItemResult reads the outcome off the two quantities rather than taking a reporter's word for
// it. A reporter able to assert `success` alongside a shortfall would settle a sale that still owes
// goods, so the claim and the numbers can never disagree here: there is only one of them.
func DeriveItemResult(dispensed, failed decimal.Decimal) FulfillmentAttemptItemResult {
	switch {
	case dispensed.IsZero() && failed.IsZero():
		return FulfillmentAttemptItemResultPending
	case failed.IsZero():
		return FulfillmentAttemptItemResultSuccess
	case dispensed.IsZero():
		return FulfillmentAttemptItemResultFailure
	}
	return FulfillmentAttemptItemResultPartial
}

// AssertQuantitiesBalance states the invariant of one reported item: what was tried is exactly what
// came out plus what did not. A result that does not balance is describing something other than the
// attempt it claims to answer — quantity that vanished, or quantity that appeared from nowhere — and
// is refused rather than reconciled, because nothing here could say which of the three numbers lied.
func (this SalesFulfillmentAttemptItem) AssertQuantitiesBalance() bool {
	attempted := attemptQty(this.GetAttemptedQty())
	dispensed := attemptQty(this.GetDispensedQty())
	failed := attemptQty(this.GetFailedQty())

	if attempted.IsNegative() || dispensed.IsNegative() || failed.IsNegative() {
		return false
	}
	return dispensed.Add(failed).Equal(attempted)
}

// DispensedQuantity is what actually reached the customer, the only quantity that raises what a sale
// has delivered.
func (this SalesFulfillmentAttemptItem) DispensedQuantity() decimal.Decimal {
	return attemptQty(this.GetDispensedQty())
}

func attemptQty(value *decimal.Decimal) decimal.Decimal {
	if value == nil {
		return decimal.Zero
	}
	return *value
}

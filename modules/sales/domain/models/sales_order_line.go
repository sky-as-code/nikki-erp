package models

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	SalesOrderLineSchemaName = "sales_order_line"

	SalesOrderLineFieldId                       = "id"
	SalesOrderLineFieldOrgId                    = "org_id"
	SalesOrderLineFieldSalesOrderId             = "sales_order_id"
	SalesOrderLineFieldLineNumber               = "line_number"
	SalesOrderLineFieldLineType                 = "line_type"
	SalesOrderLineFieldProductVariantId         = "product_variant_id"
	SalesOrderLineFieldProductCodeSnapshot      = "product_code_snapshot"
	SalesOrderLineFieldProductNameSnapshot      = "product_name_snapshot"
	SalesOrderLineFieldUomId                    = "uom_id"
	SalesOrderLineFieldOrderedQuantity          = "ordered_quantity"
	SalesOrderLineFieldRequiresFulfillment      = "requires_fulfillment"
	SalesOrderLineFieldFulfilledQuantity        = "fulfilled_quantity"
	SalesOrderLineFieldReturnedQuantity         = "returned_quantity"
	SalesOrderLineFieldBaseUnitPrice            = "base_unit_price"
	SalesOrderLineFieldEffectiveUnitPrice       = "effective_unit_price"
	SalesOrderLineFieldGrossAmount              = "gross_amount"
	SalesOrderLineFieldDiscountAmount           = "discount_amount"
	SalesOrderLineFieldNetAmount                = "net_amount"
	SalesOrderLineFieldTaxRateSnapshot          = "tax_rate_snapshot"
	SalesOrderLineFieldTaxAmount                = "tax_amount"
	SalesOrderLineFieldFinalAmount              = "final_amount"
	SalesOrderLineFieldPricingSource            = "pricing_source"
	SalesOrderLineFieldSourcePromotionProgramId = "source_promotion_program_id"
	SalesOrderLineFieldSalesComboId             = "sales_combo_id"

	SalesOrderLineEdgeSalesOrder = "sales_order"
)

// SnapshotFields are the columns frozen at confirmation. Any new field capturing how the world
// looked at the time of sale belongs here; forgetting to add it is how a snapshot silently becomes
// editable.
var SnapshotFields = []string{
	SalesOrderLineFieldProductVariantId,
	SalesOrderLineFieldProductCodeSnapshot,
	SalesOrderLineFieldProductNameSnapshot,
	SalesOrderLineFieldUomId,
	SalesOrderLineFieldBaseUnitPrice,
	SalesOrderLineFieldEffectiveUnitPrice,
	SalesOrderLineFieldTaxRateSnapshot,
}

//go:embed sales_order_line.json
var salesOrderLineSchemaJson string

func SalesOrderLineSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesOrderLineSchemaJson)
}

// SalesOrderLine is one thing sold on one order. It carries three quantities — ordered, fulfilled,
// returned — because partial fulfilment is normal and a refund is computed from the difference. The
// snapshot fields exist so a product renamed or repriced later cannot change what a receipt issued
// today says was sold.
type SalesOrderLine struct {
	basemodel.DynamicModelBase
}

func NewSalesOrderLine() *SalesOrderLine {
	return &SalesOrderLine{basemodel.NewDynamicModel()}
}

func NewSalesOrderLineFrom(src dmodel.DynamicFields) *SalesOrderLine {
	return &SalesOrderLine{basemodel.NewDynamicModel(src)}
}

func (this SalesOrderLine) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this *SalesOrderLine) SetId(id *model.Id) {
	this.GetFieldData().SetModelId(basemodel.FieldId, id)
}

func (this SalesOrderLine) GetSalesOrderId() *model.Id {
	return this.GetFieldData().GetModelId(SalesOrderLineFieldSalesOrderId)
}

func (this *SalesOrderLine) SetSalesOrderId(id *model.Id) {
	this.GetFieldData().SetModelId(SalesOrderLineFieldSalesOrderId, id)
}

func (this SalesOrderLine) GetLineNumber() *int32 {
	return this.GetFieldData().GetInt32(SalesOrderLineFieldLineNumber)
}

func (this *SalesOrderLine) SetLineNumber(number *int32) {
	this.GetFieldData().SetInt32(SalesOrderLineFieldLineNumber, number)
}

func (this SalesOrderLine) GetLineType() *string {
	return this.GetFieldData().GetString(SalesOrderLineFieldLineType)
}

func (this *SalesOrderLine) SetLineType(lineType *string) {
	this.GetFieldData().SetString(SalesOrderLineFieldLineType, lineType)
}

func (this SalesOrderLine) GetProductVariantId() *model.Id {
	return this.GetFieldData().GetModelId(SalesOrderLineFieldProductVariantId)
}

func (this *SalesOrderLine) SetProductVariantId(id *model.Id) {
	this.GetFieldData().SetModelId(SalesOrderLineFieldProductVariantId, id)
}

func (this SalesOrderLine) GetProductCodeSnapshot() *string {
	return this.GetFieldData().GetString(SalesOrderLineFieldProductCodeSnapshot)
}

func (this *SalesOrderLine) SetProductCodeSnapshot(code *string) {
	this.GetFieldData().SetString(SalesOrderLineFieldProductCodeSnapshot, code)
}

func (this SalesOrderLine) GetProductNameSnapshot() *string {
	return this.GetFieldData().GetString(SalesOrderLineFieldProductNameSnapshot)
}

func (this *SalesOrderLine) SetProductNameSnapshot(name *string) {
	this.GetFieldData().SetString(SalesOrderLineFieldProductNameSnapshot, name)
}

func (this SalesOrderLine) GetUomId() *model.Id {
	return this.GetFieldData().GetModelId(SalesOrderLineFieldUomId)
}

func (this *SalesOrderLine) SetUomId(id *model.Id) {
	this.GetFieldData().SetModelId(SalesOrderLineFieldUomId, id)
}

func (this SalesOrderLine) GetOrderedQuantity() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesOrderLineFieldOrderedQuantity)
}

func (this *SalesOrderLine) SetOrderedQuantity(quantity *decimal.Decimal) {
	this.GetFieldData().SetDecimal(SalesOrderLineFieldOrderedQuantity, quantity)
}

func (this SalesOrderLine) GetFulfilledQuantity() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesOrderLineFieldFulfilledQuantity)
}

func (this *SalesOrderLine) SetFulfilledQuantity(quantity *decimal.Decimal) {
	this.GetFieldData().SetDecimal(SalesOrderLineFieldFulfilledQuantity, quantity)
}

func (this SalesOrderLine) GetReturnedQuantity() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesOrderLineFieldReturnedQuantity)
}

func (this *SalesOrderLine) SetReturnedQuantity(quantity *decimal.Decimal) {
	this.GetFieldData().SetDecimal(SalesOrderLineFieldReturnedQuantity, quantity)
}

func (this SalesOrderLine) GetBaseUnitPrice() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesOrderLineFieldBaseUnitPrice)
}

func (this *SalesOrderLine) SetBaseUnitPrice(price *decimal.Decimal) {
	this.GetFieldData().SetDecimal(SalesOrderLineFieldBaseUnitPrice, price)
}

func (this SalesOrderLine) GetEffectiveUnitPrice() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesOrderLineFieldEffectiveUnitPrice)
}

func (this *SalesOrderLine) SetEffectiveUnitPrice(price *decimal.Decimal) {
	this.GetFieldData().SetDecimal(SalesOrderLineFieldEffectiveUnitPrice, price)
}

func (this SalesOrderLine) GetGrossAmount() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesOrderLineFieldGrossAmount)
}

func (this *SalesOrderLine) SetGrossAmount(amount *decimal.Decimal) {
	this.GetFieldData().SetDecimal(SalesOrderLineFieldGrossAmount, amount)
}

func (this SalesOrderLine) GetDiscountAmount() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesOrderLineFieldDiscountAmount)
}

func (this *SalesOrderLine) SetDiscountAmount(amount *decimal.Decimal) {
	this.GetFieldData().SetDecimal(SalesOrderLineFieldDiscountAmount, amount)
}

func (this SalesOrderLine) GetNetAmount() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesOrderLineFieldNetAmount)
}

func (this *SalesOrderLine) SetNetAmount(amount *decimal.Decimal) {
	this.GetFieldData().SetDecimal(SalesOrderLineFieldNetAmount, amount)
}

func (this SalesOrderLine) GetTaxRateSnapshot() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesOrderLineFieldTaxRateSnapshot)
}

func (this *SalesOrderLine) SetTaxRateSnapshot(rate *decimal.Decimal) {
	this.GetFieldData().SetDecimal(SalesOrderLineFieldTaxRateSnapshot, rate)
}

func (this SalesOrderLine) GetTaxAmount() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesOrderLineFieldTaxAmount)
}

func (this *SalesOrderLine) SetTaxAmount(amount *decimal.Decimal) {
	this.GetFieldData().SetDecimal(SalesOrderLineFieldTaxAmount, amount)
}

func (this SalesOrderLine) GetFinalAmount() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesOrderLineFieldFinalAmount)
}

func (this *SalesOrderLine) SetFinalAmount(amount *decimal.Decimal) {
	this.GetFieldData().SetDecimal(SalesOrderLineFieldFinalAmount, amount)
}

func (this SalesOrderLine) GetPricingSource() *string {
	return this.GetFieldData().GetString(SalesOrderLineFieldPricingSource)
}

func (this *SalesOrderLine) SetPricingSource(source *string) {
	this.GetFieldData().SetString(SalesOrderLineFieldPricingSource, source)
}

func (this SalesOrderLine) GetSourcePromotionProgramId() *model.Id {
	return this.GetFieldData().GetModelId(SalesOrderLineFieldSourcePromotionProgramId)
}

func (this *SalesOrderLine) SetSourcePromotionProgramId(id *model.Id) {
	this.GetFieldData().SetModelId(SalesOrderLineFieldSourcePromotionProgramId, id)
}

func (this SalesOrderLine) GetSalesComboId() *model.Id {
	return this.GetFieldData().GetModelId(SalesOrderLineFieldSalesComboId)
}

func (this *SalesOrderLine) SetSalesComboId(id *model.Id) {
	this.GetFieldData().SetModelId(SalesOrderLineFieldSalesComboId, id)
}

// RemainingFulfillable is how much of this line has still to be handed over. Clamped at zero rather
// than going negative, so a broken quantity invariant does not propagate into the next fulfilment
// request.
func (this SalesOrderLine) RemainingFulfillable() decimal.Decimal {
	ordered := decimalOrZero(this.GetOrderedQuantity())
	fulfilled := decimalOrZero(this.GetFulfilledQuantity())
	remaining := ordered.Sub(fulfilled)
	if remaining.IsNegative() {
		return decimal.Zero
	}
	return remaining
}

// RemainingReturnable is how much of this line may still come back. Measured against what was
// fulfilled, not ordered: a customer cannot return what was never handed over.
func (this SalesOrderLine) RemainingReturnable() decimal.Decimal {
	fulfilled := decimalOrZero(this.GetFulfilledQuantity())
	returned := decimalOrZero(this.GetReturnedQuantity())
	remaining := fulfilled.Sub(returned)
	if remaining.IsNegative() {
		return decimal.Zero
	}
	return remaining
}

// QuantitiesAreConsistent reports whether this line satisfies 0 < ordered,
// fulfilled <= ordered, returned <= fulfilled. The framework declares no CHECK constraints, so this
// is the only place the invariant is enforced, for every path that produces a line.
func (this SalesOrderLine) QuantitiesAreConsistent() bool {
	ordered := decimalOrZero(this.GetOrderedQuantity())
	if !ordered.IsPositive() {
		return false
	}
	fulfilled := decimalOrZero(this.GetFulfilledQuantity())
	if fulfilled.IsNegative() || fulfilled.GreaterThan(ordered) {
		return false
	}
	returned := decimalOrZero(this.GetReturnedQuantity())
	return !returned.IsNegative() && !returned.GreaterThan(fulfilled)
}

// IsFullyFulfilled reports whether everything ordered has been handed over.
func (this SalesOrderLine) IsFullyFulfilled() bool {
	ordered := decimalOrZero(this.GetOrderedQuantity())
	return ordered.IsPositive() && decimalOrZero(this.GetFulfilledQuantity()).GreaterThanOrEqual(ordered)
}

// decimalOrZero treats an absent amount as zero. Every quantity and money field on a line defaults
// to "0", so a nil means a partial field set was read, not an unknown value.
func decimalOrZero(value *decimal.Decimal) decimal.Decimal {
	if value == nil {
		return decimal.Zero
	}
	return *value
}

package models

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// Quotations: an offer, and the lines it offers. A quotation is not a status on sales_orders,
// because a quotation that never converts would leave a hole in the order sequence that fiscal and
// accounting systems read. The lines are derived from sales_order_lines, not shared: a quotation
// moves no goods, so there is no fulfilled/returned quantity, and one unit_price rather than base
// and effective — the explanation is rebuilt when the engine runs again at conversion.

const (
	SalesQuotationSchemaName = "sales_quotation"

	SalesQuotationFieldId             = basemodel.FieldId
	SalesQuotationFieldOrgId          = basemodel.FieldOrgId
	SalesQuotationFieldNumber         = "quotation_number"
	SalesQuotationFieldSalesChannelId = "sales_channel_id"
	SalesQuotationFieldSalesPointId   = "sales_point_id"
	SalesQuotationFieldCustomerRef    = "customer_reference"
	SalesQuotationFieldCurrencyCode   = "currency_code"
	SalesQuotationFieldStatus         = "status"
	SalesQuotationFieldValidUntil     = "valid_until"
	SalesQuotationFieldSubtotal       = "subtotal"
	SalesQuotationFieldDiscountTotal  = "discount_total"
	SalesQuotationFieldTaxTotal       = "tax_total"
	SalesQuotationFieldGrandTotal     = "grand_total"
	SalesQuotationFieldConvertedOrder = "converted_sales_order_id"
	SalesQuotationFieldSentAt         = "sent_at"
	SalesQuotationFieldAcceptedAt     = "accepted_at"
	SalesQuotationFieldCancelledAt    = "cancelled_at"

	SalesQuotationEdgeSalesChannel = "sales_channel"
)

const (
	SalesQuotationLineSchemaName = "sales_quotation_line"

	SalesQuotationLineFieldId             = basemodel.FieldId
	SalesQuotationLineFieldOrgId          = basemodel.FieldOrgId
	SalesQuotationLineFieldQuotationId    = "sales_quotation_id"
	SalesQuotationLineFieldLineNumber     = "line_number"
	SalesQuotationLineFieldVariantId      = "product_variant_id"
	SalesQuotationLineFieldProductCode    = "product_code_snapshot"
	SalesQuotationLineFieldProductName    = "product_name_snapshot"
	SalesQuotationLineFieldUomId          = "uom_id"
	SalesQuotationLineFieldQuantity       = "quantity"
	SalesQuotationLineFieldUnitPrice      = "unit_price"
	SalesQuotationLineFieldDiscountAmount = "discount_amount"
	SalesQuotationLineFieldNetAmount      = "net_amount"
	SalesQuotationLineFieldTaxAmount      = "tax_amount"
	SalesQuotationLineFieldFinalAmount    = "final_amount"

	SalesQuotationLineEdgeQuotation = "sales_quotation"
)

//go:embed sales_quotation.json
var salesQuotationSchemaJson string

func SalesQuotationSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesQuotationSchemaJson)
}

//go:embed sales_quotation_line.json
var salesQuotationLineSchemaJson string

func SalesQuotationLineSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesQuotationLineSchemaJson)
}

// SalesQuotation is one offer.
type SalesQuotation struct {
	basemodel.DynamicModelBase
}

func NewSalesQuotationFrom(src dmodel.DynamicFields) *SalesQuotation {
	return &SalesQuotation{basemodel.NewDynamicModel(src)}
}

func (this SalesQuotation) GetId() *model.Id {
	return this.GetFieldData().GetModelId(SalesQuotationFieldId)
}

func (this SalesQuotation) GetQuotationNumber() *string {
	return this.GetFieldData().GetString(SalesQuotationFieldNumber)
}

func (this SalesQuotation) GetStatus() *string {
	return this.GetFieldData().GetString(SalesQuotationFieldStatus)
}

func (this *SalesQuotation) SetStatus(status *string) {
	this.GetFieldData().SetString(SalesQuotationFieldStatus, status)
}

func (this SalesQuotation) GetCurrencyCode() *string {
	return this.GetFieldData().GetString(SalesQuotationFieldCurrencyCode)
}

func (this SalesQuotation) GetValidUntil() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(SalesQuotationFieldValidUntil)
}

func (this SalesQuotation) GetGrandTotal() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesQuotationFieldGrandTotal)
}

func (this SalesQuotation) GetConvertedSalesOrderId() *model.Id {
	return this.GetFieldData().GetModelId(SalesQuotationFieldConvertedOrder)
}

// IsConverted reports whether this quotation already became an order. This is conversion's
// idempotency check: a second accept must return the first order, not create a second delivery and
// invoice.
func (this SalesQuotation) IsConverted() bool {
	return this.GetConvertedSalesOrderId() != nil
}

// IsOffered reports whether the customer has been shown this quotation. A sent quotation may still
// be edited, but that changes an offer already made, so edits check this and are audited.
func (this SalesQuotation) IsOffered() bool {
	status := this.GetStatus()
	if status == nil {
		return false
	}
	switch SalesQuotationStatus(*status) {
	case SalesQuotationStatusSent, SalesQuotationStatusAccepted:
		return true
	}
	return false
}

// SalesQuotationLine is one line of an offer.
type SalesQuotationLine struct {
	basemodel.DynamicModelBase
}

func NewSalesQuotationLineFrom(src dmodel.DynamicFields) *SalesQuotationLine {
	return &SalesQuotationLine{basemodel.NewDynamicModel(src)}
}

func (this SalesQuotationLine) GetId() *model.Id {
	return this.GetFieldData().GetModelId(SalesQuotationLineFieldId)
}

func (this SalesQuotationLine) GetQuantity() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesQuotationLineFieldQuantity)
}

func (this SalesQuotationLine) GetFinalAmount() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesQuotationLineFieldFinalAmount)
}

// Manual price overrides. Stored rather than applied-and-forgotten because repricing deletes the
// whole adjustment chain and rewrites it from engine output, and confirm reprices unconditionally —
// an override written into sales_order_adjustments would vanish before the sale completed. These
// rows are engine input, replayed on every calculation. The base price is never rewritten.

const (
	SalesManualDiscountSchemaName = "sales_manual_discount"

	SalesManualDiscountFieldId             = basemodel.FieldId
	SalesManualDiscountFieldOrgId          = basemodel.FieldOrgId
	SalesManualDiscountFieldSalesOrderId   = "sales_order_id"
	SalesManualDiscountFieldOrderLineId    = "sales_order_line_id"
	SalesManualDiscountFieldAmount         = "discount_amount"
	SalesManualDiscountFieldReason         = "reason"
	SalesManualDiscountFieldGrantedBy      = "granted_by"
	SalesManualDiscountFieldOriginalAmount = "original_amount"

	SalesManualDiscountEdgeSalesOrder = "sales_order"
)

//go:embed sales_manual_discount.json
var salesManualDiscountSchemaJson string

func SalesManualDiscountSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesManualDiscountSchemaJson)
}

// SalesManualDiscount is one operator override.
type SalesManualDiscount struct {
	basemodel.DynamicModelBase
}

func NewSalesManualDiscountFrom(src dmodel.DynamicFields) *SalesManualDiscount {
	return &SalesManualDiscount{basemodel.NewDynamicModel(src)}
}

func (this SalesManualDiscount) GetId() *model.Id {
	return this.GetFieldData().GetModelId(SalesManualDiscountFieldId)
}

func (this SalesManualDiscount) GetDiscountAmount() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesManualDiscountFieldAmount)
}

func (this SalesManualDiscount) GetReason() *string {
	return this.GetFieldData().GetString(SalesManualDiscountFieldReason)
}

// IsOrderLevel reports whether this override applies to the basket rather than to one line. An
// order-level override is spread proportionally across the lines; a line-level one lands where it
// was granted.
func (this SalesManualDiscount) IsOrderLevel() bool {
	return this.GetFieldData().GetModelId(SalesManualDiscountFieldOrderLineId) == nil
}

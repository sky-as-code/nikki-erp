package models

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	SalesOrderAdjustmentSchemaName = "sales_order_adjustment"

	SalesOrderAdjustmentFieldId               = "id"
	SalesOrderAdjustmentFieldOrgId            = "org_id"
	SalesOrderAdjustmentFieldSalesOrderId     = "sales_order_id"
	SalesOrderAdjustmentFieldSalesOrderLineId = "sales_order_line_id"
	SalesOrderAdjustmentFieldSequence         = "sequence"
	SalesOrderAdjustmentFieldAdjustmentType   = "adjustment_type"
	SalesOrderAdjustmentFieldSourceType       = "source_type"
	SalesOrderAdjustmentFieldSourceId         = "source_id"
	SalesOrderAdjustmentFieldDescription      = "description"
	SalesOrderAdjustmentFieldBaseAmount       = "base_amount"
	SalesOrderAdjustmentFieldAdjustmentAmount = "adjustment_amount"
	SalesOrderAdjustmentFieldSalesReturnId    = "sales_return_id"

	SalesOrderAdjustmentEdgeSalesOrder = "sales_order"
)

//go:embed sales_order_adjustment.json
var salesOrderAdjustmentSchemaJson string

func SalesOrderAdjustmentSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesOrderAdjustmentSchemaJson)
}

// SalesOrderAdjustment is one step of the pricing calculation, kept so the price can be explained.
//
// BR 13 requires the pricing engine to return a LIST of adjustments rather than a total, and this
// is where that list is stored. The reason is that discounts do not commute: a percentage applied
// before a fixed amount gives a different answer than the reverse, so a stored total alone could
// never be reproduced. The sequence is what makes the replay exact.
//
// This is the PRICING trail. The document trail - who confirmed the order, when it was cancelled -
// is sales_order_events, and the two are deliberately separate: one answers "why is this number
// what it is", the other "what happened to this order".
type SalesOrderAdjustment struct {
	basemodel.DynamicModelBase
}

func NewSalesOrderAdjustment() *SalesOrderAdjustment {
	return &SalesOrderAdjustment{basemodel.NewDynamicModel()}
}

func NewSalesOrderAdjustmentFrom(src dmodel.DynamicFields) *SalesOrderAdjustment {
	return &SalesOrderAdjustment{basemodel.NewDynamicModel(src)}
}

func (this SalesOrderAdjustment) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this *SalesOrderAdjustment) SetId(id *model.Id) {
	this.GetFieldData().SetModelId(basemodel.FieldId, id)
}

func (this SalesOrderAdjustment) GetSalesOrderId() *model.Id {
	return this.GetFieldData().GetModelId(SalesOrderAdjustmentFieldSalesOrderId)
}

func (this *SalesOrderAdjustment) SetSalesOrderId(id *model.Id) {
	this.GetFieldData().SetModelId(SalesOrderAdjustmentFieldSalesOrderId, id)
}

func (this SalesOrderAdjustment) GetSalesOrderLineId() *model.Id {
	return this.GetFieldData().GetModelId(SalesOrderAdjustmentFieldSalesOrderLineId)
}

func (this *SalesOrderAdjustment) SetSalesOrderLineId(id *model.Id) {
	this.GetFieldData().SetModelId(SalesOrderAdjustmentFieldSalesOrderLineId, id)
}

func (this SalesOrderAdjustment) GetSequence() *int32 {
	return this.GetFieldData().GetInt32(SalesOrderAdjustmentFieldSequence)
}

func (this *SalesOrderAdjustment) SetSequence(sequence *int32) {
	this.GetFieldData().SetInt32(SalesOrderAdjustmentFieldSequence, sequence)
}

func (this SalesOrderAdjustment) GetAdjustmentType() *string {
	return this.GetFieldData().GetString(SalesOrderAdjustmentFieldAdjustmentType)
}

func (this *SalesOrderAdjustment) SetAdjustmentType(adjustmentType *string) {
	this.GetFieldData().SetString(SalesOrderAdjustmentFieldAdjustmentType, adjustmentType)
}

func (this SalesOrderAdjustment) GetSourceType() *string {
	return this.GetFieldData().GetString(SalesOrderAdjustmentFieldSourceType)
}

func (this *SalesOrderAdjustment) SetSourceType(sourceType *string) {
	this.GetFieldData().SetString(SalesOrderAdjustmentFieldSourceType, sourceType)
}

func (this SalesOrderAdjustment) GetSourceId() *model.Id {
	return this.GetFieldData().GetModelId(SalesOrderAdjustmentFieldSourceId)
}

func (this *SalesOrderAdjustment) SetSourceId(id *model.Id) {
	this.GetFieldData().SetModelId(SalesOrderAdjustmentFieldSourceId, id)
}

func (this SalesOrderAdjustment) GetDescription() *string {
	return this.GetFieldData().GetString(SalesOrderAdjustmentFieldDescription)
}

func (this *SalesOrderAdjustment) SetDescription(description *string) {
	this.GetFieldData().SetString(SalesOrderAdjustmentFieldDescription, description)
}

func (this SalesOrderAdjustment) GetBaseAmount() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesOrderAdjustmentFieldBaseAmount)
}

func (this *SalesOrderAdjustment) SetBaseAmount(amount *decimal.Decimal) {
	this.GetFieldData().SetDecimal(SalesOrderAdjustmentFieldBaseAmount, amount)
}

func (this SalesOrderAdjustment) GetAdjustmentAmount() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesOrderAdjustmentFieldAdjustmentAmount)
}

func (this *SalesOrderAdjustment) SetAdjustmentAmount(amount *decimal.Decimal) {
	this.GetFieldData().SetDecimal(SalesOrderAdjustmentFieldAdjustmentAmount, amount)
}

func (this SalesOrderAdjustment) GetSalesReturnId() *model.Id {
	return this.GetFieldData().GetModelId(SalesOrderAdjustmentFieldSalesReturnId)
}

func (this *SalesOrderAdjustment) SetSalesReturnId(id *model.Id) {
	this.GetFieldData().SetModelId(SalesOrderAdjustmentFieldSalesReturnId, id)
}

// IsOrderLevel reports whether this adjustment applies to the order as a whole rather than to one
// line.
//
// The distinction matters when replaying: an order-level discount is allocated across the lines, so
// it appears once here and again as each line's share of discount_amount. Counting both would
// double it.
func (this SalesOrderAdjustment) IsOrderLevel() bool {
	return this.GetSalesOrderLineId() == nil
}

// IsReturnClawback reports whether this adjustment was raised by a return rather than by the
// original sale (D-12).
func (this SalesOrderAdjustment) IsReturnClawback() bool {
	return this.GetSalesReturnId() != nil
}

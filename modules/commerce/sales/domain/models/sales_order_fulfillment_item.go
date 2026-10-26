package models

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	SalesOrderFulfillmentItemSchemaName = "sales_order_fulfillment_item"

	SalesOrderFulfillmentItemFieldId                      = "id"
	SalesOrderFulfillmentItemFieldOrgId                   = "org_id"
	SalesOrderFulfillmentItemFieldFulfillmentId           = "fulfillment_id"
	SalesOrderFulfillmentItemFieldSalesOrderLineId        = "sales_order_line_id"
	SalesOrderFulfillmentItemFieldProductVariantId        = "product_variant_id"
	SalesOrderFulfillmentItemFieldUomId                   = "uom_id"
	SalesOrderFulfillmentItemFieldOrderedQty              = "ordered_qty"
	SalesOrderFulfillmentItemFieldFulfilledQty            = "fulfilled_qty"
	SalesOrderFulfillmentItemFieldRefundedQty             = "refunded_qty"
	SalesOrderFulfillmentItemFieldSourceLocationId        = "source_location_id"
	SalesOrderFulfillmentItemFieldInventoryReservationRef = "inventory_reservation_ref"
	SalesOrderFulfillmentItemFieldItemStatus              = "item_status"
	SalesOrderFulfillmentItemFieldRemainingQty            = "remaining_qty"
)

//go:embed sales_order_fulfillment_item.json
var salesOrderFulfillmentItemSchemaJson string

func SalesOrderFulfillmentItemSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesOrderFulfillmentItemSchemaJson)
}

type SalesOrderFulfillmentItem struct {
	basemodel.DynamicModelBase
}

func NewSalesOrderFulfillmentItem() *SalesOrderFulfillmentItem {
	return &SalesOrderFulfillmentItem{basemodel.NewDynamicModel()}
}

func NewSalesOrderFulfillmentItemFrom(src dmodel.DynamicFields) *SalesOrderFulfillmentItem {
	return &SalesOrderFulfillmentItem{basemodel.NewDynamicModel(src)}
}

func (this SalesOrderFulfillmentItem) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this *SalesOrderFulfillmentItem) SetId(id *model.Id) {
	this.GetFieldData().SetModelId(basemodel.FieldId, id)
}

func (this SalesOrderFulfillmentItem) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(SalesOrderFulfillmentItemFieldOrgId)
}

func (this *SalesOrderFulfillmentItem) SetOrgId(value *model.Id) {
	this.GetFieldData().SetModelId(SalesOrderFulfillmentItemFieldOrgId, value)
}

func (this SalesOrderFulfillmentItem) GetFulfillmentId() *model.Id {
	return this.GetFieldData().GetModelId(SalesOrderFulfillmentItemFieldFulfillmentId)
}

func (this *SalesOrderFulfillmentItem) SetFulfillmentId(value *model.Id) {
	this.GetFieldData().SetModelId(SalesOrderFulfillmentItemFieldFulfillmentId, value)
}

func (this SalesOrderFulfillmentItem) GetSalesOrderLineId() *model.Id {
	return this.GetFieldData().GetModelId(SalesOrderFulfillmentItemFieldSalesOrderLineId)
}

func (this *SalesOrderFulfillmentItem) SetSalesOrderLineId(value *model.Id) {
	this.GetFieldData().SetModelId(SalesOrderFulfillmentItemFieldSalesOrderLineId, value)
}

func (this SalesOrderFulfillmentItem) GetProductVariantId() *model.Id {
	return this.GetFieldData().GetModelId(SalesOrderFulfillmentItemFieldProductVariantId)
}

func (this *SalesOrderFulfillmentItem) SetProductVariantId(value *model.Id) {
	this.GetFieldData().SetModelId(SalesOrderFulfillmentItemFieldProductVariantId, value)
}

func (this SalesOrderFulfillmentItem) GetUomId() *model.Id {
	return this.GetFieldData().GetModelId(SalesOrderFulfillmentItemFieldUomId)
}

func (this *SalesOrderFulfillmentItem) SetUomId(value *model.Id) {
	this.GetFieldData().SetModelId(SalesOrderFulfillmentItemFieldUomId, value)
}

func (this SalesOrderFulfillmentItem) GetOrderedQty() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesOrderFulfillmentItemFieldOrderedQty)
}

func (this *SalesOrderFulfillmentItem) SetOrderedQty(value *decimal.Decimal) {
	this.GetFieldData().SetDecimal(SalesOrderFulfillmentItemFieldOrderedQty, value)
}

func (this SalesOrderFulfillmentItem) GetFulfilledQty() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesOrderFulfillmentItemFieldFulfilledQty)
}

func (this *SalesOrderFulfillmentItem) SetFulfilledQty(value *decimal.Decimal) {
	this.GetFieldData().SetDecimal(SalesOrderFulfillmentItemFieldFulfilledQty, value)
}

func (this SalesOrderFulfillmentItem) GetRefundedQty() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesOrderFulfillmentItemFieldRefundedQty)
}

func (this *SalesOrderFulfillmentItem) SetRefundedQty(value *decimal.Decimal) {
	this.GetFieldData().SetDecimal(SalesOrderFulfillmentItemFieldRefundedQty, value)
}

func (this SalesOrderFulfillmentItem) GetSourceLocationId() *model.Id {
	return this.GetFieldData().GetModelId(SalesOrderFulfillmentItemFieldSourceLocationId)
}

func (this *SalesOrderFulfillmentItem) SetSourceLocationId(value *model.Id) {
	this.GetFieldData().SetModelId(SalesOrderFulfillmentItemFieldSourceLocationId, value)
}

func (this SalesOrderFulfillmentItem) GetInventoryReservationRef() *string {
	return this.GetFieldData().GetString(SalesOrderFulfillmentItemFieldInventoryReservationRef)
}

func (this *SalesOrderFulfillmentItem) SetInventoryReservationRef(value *string) {
	this.GetFieldData().SetString(SalesOrderFulfillmentItemFieldInventoryReservationRef, value)
}

func (this SalesOrderFulfillmentItem) GetItemStatus() *string {
	return this.GetFieldData().GetString(SalesOrderFulfillmentItemFieldItemStatus)
}

func (this *SalesOrderFulfillmentItem) SetItemStatus(value *string) {
	this.GetFieldData().SetString(SalesOrderFulfillmentItemFieldItemStatus, value)
}

func (this SalesOrderFulfillmentItem) GetRemainingQty() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesOrderFulfillmentItemFieldRemainingQty)
}

// zeroQty is the value a nil quantity stands for. The columns default to 0, so a nil here means the
// record was read without the field rather than that nothing was delivered; both answer the same
// arithmetic, and a nil-check at every call site would say the same thing five times.
func zeroQty(value *decimal.Decimal) decimal.Decimal {
	if value == nil {
		return decimal.Zero
	}
	return *value
}

// RemainingQuantity is what the customer is still owed. The column of the same name is computed by
// the engine for reading and filtering; this recomputes it in Go for the services that decide on a
// record they are holding mid-transaction, where the stored projection has not been refreshed yet.
func (this SalesOrderFulfillmentItem) RemainingQuantity() decimal.Decimal {
	remaining := zeroQty(this.GetOrderedQty()).
		Sub(zeroQty(this.GetFulfilledQty())).
		Sub(zeroQty(this.GetRefundedQty()))
	if remaining.IsNegative() {
		return decimal.Zero
	}
	return remaining
}

// AssertQuantitiesValid states the invariant that makes every quantity above meaningful: nothing is
// negative, and the two settled outcomes together never exceed what was ordered. Violating it means
// a customer was credited for more than they bought — as goods, as money, or as both — so it is
// checked at every write rather than trusted, and returns the offending field so a caller can say
// which one broke.
func (this SalesOrderFulfillmentItem) AssertQuantitiesValid() (violatedField string, ok bool) {
	ordered := zeroQty(this.GetOrderedQty())
	fulfilled := zeroQty(this.GetFulfilledQty())
	refunded := zeroQty(this.GetRefundedQty())

	switch {
	case ordered.IsNegative():
		return SalesOrderFulfillmentItemFieldOrderedQty, false
	case fulfilled.IsNegative():
		return SalesOrderFulfillmentItemFieldFulfilledQty, false
	case refunded.IsNegative():
		return SalesOrderFulfillmentItemFieldRefundedQty, false
	case fulfilled.Add(refunded).GreaterThan(ordered):
		return SalesOrderFulfillmentItemFieldFulfilledQty, false
	}
	return "", true
}

// IsSettled reports that this item owes nothing further, whether the customer got the goods, their
// money, or some of each. It is the per-item half of the completion rule, which turns on remaining
// quantity alone and never on how any refund ended.
func (this SalesOrderFulfillmentItem) IsSettled() bool {
	return this.RemainingQuantity().IsZero()
}

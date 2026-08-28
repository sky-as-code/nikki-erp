package models

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	PurchaseOrderLineSchemaName = "purchase_order_line"

	PurchaseOrderLineFieldId                   = basemodel.FieldId
	PurchaseOrderLineFieldEtag                 = basemodel.FieldEtag
	PurchaseOrderLineFieldOrgId                = basemodel.FieldOrgId
	PurchaseOrderLineFieldPurchaseOrderId      = "purchase_order_id"
	PurchaseOrderLineFieldSequence             = "sequence"
	PurchaseOrderLineFieldLineType             = "line_type"
	PurchaseOrderLineFieldProductVariantId     = "product_variant_id"
	PurchaseOrderLineFieldDescription          = "description"
	PurchaseOrderLineFieldQuantity             = "quantity"
	PurchaseOrderLineFieldUomId                = "uom_id"
	PurchaseOrderLineFieldInventoryQuantity    = "inventory_quantity"
	PurchaseOrderLineFieldUnitPrice            = "unit_price"
	PurchaseOrderLineFieldVendorProductPriceId = "vendor_product_price_id"
	PurchaseOrderLineFieldResolvedUnitPrice    = "resolved_unit_price"
	PurchaseOrderLineFieldDiscountPercent      = "discount_percent"
	PurchaseOrderLineFieldExpectedArrival      = "expected_arrival"
	PurchaseOrderLineFieldSubtotal             = "subtotal"
	PurchaseOrderLineFieldTaxAmount            = "tax_amount"
	PurchaseOrderLineFieldTotal                = "total"
	PurchaseOrderLineEdgeOrder                 = "order"
)

//go:embed purchase_order_line.json
var purchaseOrderLineSchemaJson string

func PurchaseOrderLineSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(purchaseOrderLineSchemaJson)
}

type PurchaseOrderLine struct {
	fields dmodel.DynamicFields
}

func NewPurchaseOrderLine() *PurchaseOrderLine {
	return &PurchaseOrderLine{fields: make(dmodel.DynamicFields)}
}

func NewPurchaseOrderLineFrom(src dmodel.DynamicFields) *PurchaseOrderLine {
	return &PurchaseOrderLine{fields: src}
}

func (this PurchaseOrderLine) GetFieldData() dmodel.DynamicFields {
	return this.fields
}

func (this *PurchaseOrderLine) SetFieldData(data dmodel.DynamicFields) {
	this.fields = data
}

func (this PurchaseOrderLine) GetId() *model.Id {
	return this.fields.GetModelId(PurchaseOrderLineFieldId)
}

func (this *PurchaseOrderLine) SetId(v *model.Id) {
	this.fields.SetModelId(PurchaseOrderLineFieldId, v)
}

func (this PurchaseOrderLine) GetEtag() *model.Etag {
	return this.fields.GetEtag(PurchaseOrderLineFieldEtag)
}

func (this *PurchaseOrderLine) SetEtag(v *model.Etag) {
	this.fields.SetEtag(PurchaseOrderLineFieldEtag, v)
}

func (this PurchaseOrderLine) GetOrgId() *model.Id {
	return this.fields.GetModelId(PurchaseOrderLineFieldOrgId)
}

func (this *PurchaseOrderLine) SetOrgId(v *model.Id) {
	this.fields.SetModelId(PurchaseOrderLineFieldOrgId, v)
}

func (this PurchaseOrderLine) GetPurchaseOrderId() *model.Id {
	return this.fields.GetModelId(PurchaseOrderLineFieldPurchaseOrderId)
}

func (this *PurchaseOrderLine) SetPurchaseOrderId(v *model.Id) {
	this.fields.SetModelId(PurchaseOrderLineFieldPurchaseOrderId, v)
}

func (this PurchaseOrderLine) GetSequence() *int32 {
	return this.fields.GetInt32(PurchaseOrderLineFieldSequence)
}

func (this *PurchaseOrderLine) SetSequence(v *int32) {
	this.fields.SetInt32(PurchaseOrderLineFieldSequence, v)
}

func (this PurchaseOrderLine) GetLineType() *string {
	return this.fields.GetString(PurchaseOrderLineFieldLineType)
}

func (this *PurchaseOrderLine) SetLineType(v *string) {
	this.fields.SetString(PurchaseOrderLineFieldLineType, v)
}

func (this PurchaseOrderLine) GetProductVariantId() *model.Id {
	return this.fields.GetModelId(PurchaseOrderLineFieldProductVariantId)
}

func (this *PurchaseOrderLine) SetProductVariantId(v *model.Id) {
	this.fields.SetModelId(PurchaseOrderLineFieldProductVariantId, v)
}

func (this PurchaseOrderLine) GetDescription() *string {
	return this.fields.GetString(PurchaseOrderLineFieldDescription)
}

func (this *PurchaseOrderLine) SetDescription(v *string) {
	this.fields.SetString(PurchaseOrderLineFieldDescription, v)
}

func (this PurchaseOrderLine) GetQuantity() *decimal.Decimal {
	return this.fields.GetDecimal(PurchaseOrderLineFieldQuantity)
}

func (this *PurchaseOrderLine) SetQuantity(v *decimal.Decimal) {
	this.fields.SetDecimal(PurchaseOrderLineFieldQuantity, v)
}

func (this PurchaseOrderLine) GetUomId() *model.Id {
	return this.fields.GetModelId(PurchaseOrderLineFieldUomId)
}

func (this *PurchaseOrderLine) SetUomId(v *model.Id) {
	this.fields.SetModelId(PurchaseOrderLineFieldUomId, v)
}

func (this PurchaseOrderLine) GetInventoryQuantity() *decimal.Decimal {
	return this.fields.GetDecimal(PurchaseOrderLineFieldInventoryQuantity)
}

func (this *PurchaseOrderLine) SetInventoryQuantity(v *decimal.Decimal) {
	this.fields.SetDecimal(PurchaseOrderLineFieldInventoryQuantity, v)
}

func (this PurchaseOrderLine) GetUnitPrice() *decimal.Decimal {
	return this.fields.GetDecimal(PurchaseOrderLineFieldUnitPrice)
}

func (this *PurchaseOrderLine) SetUnitPrice(v *decimal.Decimal) {
	this.fields.SetDecimal(PurchaseOrderLineFieldUnitPrice, v)
}

func (this PurchaseOrderLine) GetVendorProductPriceId() *model.Id {
	return this.fields.GetModelId(PurchaseOrderLineFieldVendorProductPriceId)
}

func (this *PurchaseOrderLine) SetVendorProductPriceId(v *model.Id) {
	this.fields.SetModelId(PurchaseOrderLineFieldVendorProductPriceId, v)
}

func (this PurchaseOrderLine) GetResolvedUnitPrice() *decimal.Decimal {
	return this.fields.GetDecimal(PurchaseOrderLineFieldResolvedUnitPrice)
}

func (this *PurchaseOrderLine) SetResolvedUnitPrice(v *decimal.Decimal) {
	this.fields.SetDecimal(PurchaseOrderLineFieldResolvedUnitPrice, v)
}

func (this PurchaseOrderLine) GetDiscountPercent() *decimal.Decimal {
	return this.fields.GetDecimal(PurchaseOrderLineFieldDiscountPercent)
}

func (this *PurchaseOrderLine) SetDiscountPercent(v *decimal.Decimal) {
	this.fields.SetDecimal(PurchaseOrderLineFieldDiscountPercent, v)
}

func (this PurchaseOrderLine) GetExpectedArrival() *model.ModelDateTime {
	return this.fields.GetModelDateTime(PurchaseOrderLineFieldExpectedArrival)
}

func (this *PurchaseOrderLine) SetExpectedArrival(v *model.ModelDateTime) {
	this.fields.SetModelDateTime(PurchaseOrderLineFieldExpectedArrival, v)
}

func (this PurchaseOrderLine) GetSubtotal() *decimal.Decimal {
	return this.fields.GetDecimal(PurchaseOrderLineFieldSubtotal)
}

func (this *PurchaseOrderLine) SetSubtotal(v *decimal.Decimal) {
	this.fields.SetDecimal(PurchaseOrderLineFieldSubtotal, v)
}

func (this PurchaseOrderLine) GetTaxAmount() *decimal.Decimal {
	return this.fields.GetDecimal(PurchaseOrderLineFieldTaxAmount)
}

func (this *PurchaseOrderLine) SetTaxAmount(v *decimal.Decimal) {
	this.fields.SetDecimal(PurchaseOrderLineFieldTaxAmount, v)
}

func (this PurchaseOrderLine) GetTotal() *decimal.Decimal {
	return this.fields.GetDecimal(PurchaseOrderLineFieldTotal)
}

func (this *PurchaseOrderLine) SetTotal(v *decimal.Decimal) {
	this.fields.SetDecimal(PurchaseOrderLineFieldTotal, v)
}

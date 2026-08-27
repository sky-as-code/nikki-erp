package models

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	SalesOrderLineComponentSchemaName = "sales_order_line_component"

	SalesOrderLineComponentFieldId                  = "id"
	SalesOrderLineComponentFieldOrgId               = "org_id"
	SalesOrderLineComponentFieldSalesOrderLineId    = "sales_order_line_id"
	SalesOrderLineComponentFieldSequence            = "sequence"
	SalesOrderLineComponentFieldProductVariantId    = "product_variant_id"
	SalesOrderLineComponentFieldProductCodeSnapshot = "product_code_snapshot"
	SalesOrderLineComponentFieldProductNameSnapshot = "product_name_snapshot"
	SalesOrderLineComponentFieldQuantity            = "quantity"
	SalesOrderLineComponentFieldUomId               = "uom_id"
	SalesOrderLineComponentFieldAllocatedNetAmount  = "allocated_net_amount"
	SalesOrderLineComponentFieldAllocatedTaxAmount  = "allocated_tax_amount"

	SalesOrderLineComponentEdgeSalesOrderLine = "sales_order_line"
)

//go:embed sales_order_line_component.json
var salesOrderLineComponentSchemaJson string

func SalesOrderLineComponentSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesOrderLineComponentSchemaJson)
}

// SalesOrderLineComponent is one real product inside a combo.
//
// The order keeps BOTH the combo parent line and its components (BR 17), which looks like
// duplication and is not: the parent line is what the customer bought and was charged for, the
// components are what Inventory must actually hand over. Inventory fulfils real variants and has no
// concept of a virtual bundle, so without components a combo could be sold and never dispatched.
//
// The allocated amounts are what tie the two views together. Their sum across a line must equal
// that line's net amount exactly, or the receipt and the stock valuation disagree about what the
// bundle was worth.
type SalesOrderLineComponent struct {
	basemodel.DynamicModelBase
}

func NewSalesOrderLineComponent() *SalesOrderLineComponent {
	return &SalesOrderLineComponent{basemodel.NewDynamicModel()}
}

func NewSalesOrderLineComponentFrom(src dmodel.DynamicFields) *SalesOrderLineComponent {
	return &SalesOrderLineComponent{basemodel.NewDynamicModel(src)}
}

func (this SalesOrderLineComponent) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this *SalesOrderLineComponent) SetId(id *model.Id) {
	this.GetFieldData().SetModelId(basemodel.FieldId, id)
}

func (this SalesOrderLineComponent) GetSalesOrderLineId() *model.Id {
	return this.GetFieldData().GetModelId(SalesOrderLineComponentFieldSalesOrderLineId)
}

func (this *SalesOrderLineComponent) SetSalesOrderLineId(id *model.Id) {
	this.GetFieldData().SetModelId(SalesOrderLineComponentFieldSalesOrderLineId, id)
}

func (this SalesOrderLineComponent) GetSequence() *int32 {
	return this.GetFieldData().GetInt32(SalesOrderLineComponentFieldSequence)
}

func (this *SalesOrderLineComponent) SetSequence(sequence *int32) {
	this.GetFieldData().SetInt32(SalesOrderLineComponentFieldSequence, sequence)
}

func (this SalesOrderLineComponent) GetProductVariantId() *model.Id {
	return this.GetFieldData().GetModelId(SalesOrderLineComponentFieldProductVariantId)
}

func (this *SalesOrderLineComponent) SetProductVariantId(id *model.Id) {
	this.GetFieldData().SetModelId(SalesOrderLineComponentFieldProductVariantId, id)
}

func (this SalesOrderLineComponent) GetProductCodeSnapshot() *string {
	return this.GetFieldData().GetString(SalesOrderLineComponentFieldProductCodeSnapshot)
}

func (this *SalesOrderLineComponent) SetProductCodeSnapshot(code *string) {
	this.GetFieldData().SetString(SalesOrderLineComponentFieldProductCodeSnapshot, code)
}

func (this SalesOrderLineComponent) GetProductNameSnapshot() *string {
	return this.GetFieldData().GetString(SalesOrderLineComponentFieldProductNameSnapshot)
}

func (this *SalesOrderLineComponent) SetProductNameSnapshot(name *string) {
	this.GetFieldData().SetString(SalesOrderLineComponentFieldProductNameSnapshot, name)
}

func (this SalesOrderLineComponent) GetQuantity() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesOrderLineComponentFieldQuantity)
}

func (this *SalesOrderLineComponent) SetQuantity(quantity *decimal.Decimal) {
	this.GetFieldData().SetDecimal(SalesOrderLineComponentFieldQuantity, quantity)
}

func (this SalesOrderLineComponent) GetUomId() *model.Id {
	return this.GetFieldData().GetModelId(SalesOrderLineComponentFieldUomId)
}

func (this *SalesOrderLineComponent) SetUomId(id *model.Id) {
	this.GetFieldData().SetModelId(SalesOrderLineComponentFieldUomId, id)
}

func (this SalesOrderLineComponent) GetAllocatedNetAmount() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesOrderLineComponentFieldAllocatedNetAmount)
}

func (this *SalesOrderLineComponent) SetAllocatedNetAmount(amount *decimal.Decimal) {
	this.GetFieldData().SetDecimal(SalesOrderLineComponentFieldAllocatedNetAmount, amount)
}

func (this SalesOrderLineComponent) GetAllocatedTaxAmount() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesOrderLineComponentFieldAllocatedTaxAmount)
}

func (this *SalesOrderLineComponent) SetAllocatedTaxAmount(amount *decimal.Decimal) {
	this.GetFieldData().SetDecimal(SalesOrderLineComponentFieldAllocatedTaxAmount, amount)
}

// SumAllocatedNet totals the net amounts of a set of components.
//
// It exists so that the D-04 invariant has one implementation: the sum across a combo line's
// components must equal that line's net_amount EXACTLY, and every place that checks it should add
// the numbers the same way.
func SumAllocatedNet(components []dmodel.DynamicFields) decimal.Decimal {
	total := decimal.Zero
	for _, record := range components {
		total = total.Add(decimalOrZero(
			NewSalesOrderLineComponentFrom(record).GetAllocatedNetAmount()))
	}
	return total
}

// SumAllocatedTax totals the tax amounts of a set of components, for the same reason.
func SumAllocatedTax(components []dmodel.DynamicFields) decimal.Decimal {
	total := decimal.Zero
	for _, record := range components {
		total = total.Add(decimalOrZero(
			NewSalesOrderLineComponentFrom(record).GetAllocatedTaxAmount()))
	}
	return total
}

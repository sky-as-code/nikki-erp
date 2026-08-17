package models

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	InvoiceLineSchemaName = "paymentinvoice_invoice_line"

	InvoiceLineFieldId             = basemodel.FieldId
	InvoiceLineFieldInvoiceId      = "invoice_id"
	InvoiceLineFieldDescription    = "description"
	InvoiceLineFieldQuantity       = "quantity"
	InvoiceLineFieldUnitPrice      = "unit_price"
	InvoiceLineFieldTaxRatePercent = "tax_rate_percent"
	InvoiceLineFieldAmount         = "amount"
	InvoiceLineFieldOrgId          = "org_id"
)

//go:embed invoice_line.json
var invoiceLineSchemaJson string

func InvoiceLineSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(invoiceLineSchemaJson)
}

// InvoiceLine is one charged item of an invoice. Table: paymentinvoice_invoice_lines.
//
// unit_price, tax_rate_percent and amount are numeric. A tax rate is not a whole number in
// general — 8.5% is ordinary — and an integer column would truncate it rather than reject it.
//
// quantity is integer, not bigint: its bound is a million, and a wider column would reserve eight
// bytes a row to represent quantities no line can hold while telling the next reader the bound
// had not been chosen.
//
// amount is recomputed from quantity and unit_price when the invoice is issued, never accepted
// from a client, so a line cannot claim a total its own figures do not produce.
type InvoiceLine struct {
	basemodel.DynamicModelBase
}

func NewInvoiceLine() *InvoiceLine {
	return &InvoiceLine{basemodel.NewDynamicModel()}
}

func NewInvoiceLineFrom(src dmodel.DynamicFields) *InvoiceLine {
	return &InvoiceLine{basemodel.NewDynamicModel(src)}
}

func (this InvoiceLine) GetInvoiceId() *model.Id {
	return this.GetFieldData().GetModelId(InvoiceLineFieldInvoiceId)
}

func (this *InvoiceLine) SetInvoiceId(v *model.Id) {
	this.GetFieldData().SetModelId(InvoiceLineFieldInvoiceId, v)
}

func (this InvoiceLine) GetDescription() *string {
	return this.GetFieldData().GetString(InvoiceLineFieldDescription)
}

func (this *InvoiceLine) SetDescription(v *string) {
	this.GetFieldData().SetString(InvoiceLineFieldDescription, v)
}

func (this InvoiceLine) GetQuantity() *int32 {
	return this.GetFieldData().GetInt32(InvoiceLineFieldQuantity)
}

func (this *InvoiceLine) SetQuantity(v *int32) {
	this.GetFieldData().SetInt32(InvoiceLineFieldQuantity, v)
}

func (this InvoiceLine) GetUnitPrice() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(InvoiceLineFieldUnitPrice)
}

func (this *InvoiceLine) SetUnitPrice(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(InvoiceLineFieldUnitPrice, v)
}

func (this InvoiceLine) GetTaxRatePercent() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(InvoiceLineFieldTaxRatePercent)
}

func (this *InvoiceLine) SetTaxRatePercent(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(InvoiceLineFieldTaxRatePercent, v)
}

func (this InvoiceLine) GetAmount() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(InvoiceLineFieldAmount)
}

func (this *InvoiceLine) SetAmount(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(InvoiceLineFieldAmount, v)
}

func (this InvoiceLine) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(InvoiceLineFieldOrgId)
}

func (this *InvoiceLine) SetOrgId(v *model.Id) {
	this.GetFieldData().SetModelId(InvoiceLineFieldOrgId, v)
}

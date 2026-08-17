package models

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	InvoiceSchemaName = "paymentinvoice_invoice"

	InvoiceFieldId             = basemodel.FieldId
	InvoiceFieldNumber         = "number"
	InvoiceFieldStatus         = "status"
	InvoiceFieldOrderId        = "order_id"
	InvoiceFieldPartnerName    = "partner_name"
	InvoiceFieldPartnerTaxCode = "partner_tax_code"
	InvoiceFieldPartnerAddress = "partner_address"
	InvoiceFieldCurrencyId     = "currency_id"
	InvoiceFieldSubtotalAmount = "subtotal_amount"
	InvoiceFieldTaxAmount      = "tax_amount"
	InvoiceFieldTotalAmount    = "total_amount"
	InvoiceFieldIssuedAt       = "issued_at"
	InvoiceFieldNote           = "note"
	InvoiceFieldOrgId          = "org_id"
)

const (
	InvoiceStatusDraft  = "draft"
	InvoiceStatusIssued = "issued"
	InvoiceStatusPaid   = "paid"
	InvoiceStatusVoid   = "void"
)

//go:embed invoice.json
var invoiceSchemaJson string

func InvoiceSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(invoiceSchemaJson)
}

// Invoice is the accounting document raised for a sale. Table: paymentinvoice_invoices.
//
// number is nullable and unique. An invoice has no number while it is a draft, because a number
// handed out before issue would leave a gap in the sequence if the draft were abandoned; Postgres
// permits many NULLs under a UNIQUE constraint, so drafts are unconstrained while issued numbers
// stay unique. The number, the totals and issued_at are all assigned together by the issue
// action, which is why each is no_update.
//
// The partner's name, tax code and address are copied onto the invoice rather than referenced, so
// that a later change to a customer record cannot alter a document already issued.
type Invoice struct {
	basemodel.DynamicModelBase
}

func NewInvoice() *Invoice {
	return &Invoice{basemodel.NewDynamicModel()}
}

func NewInvoiceFrom(src dmodel.DynamicFields) *Invoice {
	return &Invoice{basemodel.NewDynamicModel(src)}
}

func (this Invoice) GetNumber() *string {
	return this.GetFieldData().GetString(InvoiceFieldNumber)
}

func (this *Invoice) SetNumber(v *string) {
	this.GetFieldData().SetString(InvoiceFieldNumber, v)
}

func (this Invoice) GetStatus() *string {
	return this.GetFieldData().GetString(InvoiceFieldStatus)
}

func (this *Invoice) SetStatus(v *string) {
	this.GetFieldData().SetString(InvoiceFieldStatus, v)
}

func (this Invoice) GetOrderId() *model.Id {
	return this.GetFieldData().GetModelId(InvoiceFieldOrderId)
}

func (this *Invoice) SetOrderId(v *model.Id) {
	this.GetFieldData().SetModelId(InvoiceFieldOrderId, v)
}

func (this Invoice) GetPartnerName() *string {
	return this.GetFieldData().GetString(InvoiceFieldPartnerName)
}

func (this *Invoice) SetPartnerName(v *string) {
	this.GetFieldData().SetString(InvoiceFieldPartnerName, v)
}

func (this Invoice) GetPartnerTaxCode() *string {
	return this.GetFieldData().GetString(InvoiceFieldPartnerTaxCode)
}

func (this *Invoice) SetPartnerTaxCode(v *string) {
	this.GetFieldData().SetString(InvoiceFieldPartnerTaxCode, v)
}

func (this Invoice) GetPartnerAddress() *string {
	return this.GetFieldData().GetString(InvoiceFieldPartnerAddress)
}

func (this *Invoice) SetPartnerAddress(v *string) {
	this.GetFieldData().SetString(InvoiceFieldPartnerAddress, v)
}

func (this Invoice) GetCurrencyId() *model.Id {
	return this.GetFieldData().GetModelId(InvoiceFieldCurrencyId)
}

func (this *Invoice) SetCurrencyId(v *model.Id) {
	this.GetFieldData().SetModelId(InvoiceFieldCurrencyId, v)
}

func (this Invoice) GetSubtotalAmount() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(InvoiceFieldSubtotalAmount)
}

func (this *Invoice) SetSubtotalAmount(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(InvoiceFieldSubtotalAmount, v)
}

func (this Invoice) GetTaxAmount() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(InvoiceFieldTaxAmount)
}

func (this *Invoice) SetTaxAmount(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(InvoiceFieldTaxAmount, v)
}

func (this Invoice) GetTotalAmount() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(InvoiceFieldTotalAmount)
}

func (this *Invoice) SetTotalAmount(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(InvoiceFieldTotalAmount, v)
}

func (this Invoice) GetIssuedAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(InvoiceFieldIssuedAt)
}

func (this *Invoice) SetIssuedAt(v *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(InvoiceFieldIssuedAt, v)
}

func (this Invoice) GetNote() *string {
	return this.GetFieldData().GetString(InvoiceFieldNote)
}

func (this *Invoice) SetNote(v *string) {
	this.GetFieldData().SetString(InvoiceFieldNote, v)
}

func (this Invoice) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(InvoiceFieldOrgId)
}

func (this *Invoice) SetOrgId(v *model.Id) {
	this.GetFieldData().SetModelId(InvoiceFieldOrgId, v)
}

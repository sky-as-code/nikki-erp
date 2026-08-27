package models

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	SalesOrderSchemaName = "sales_order"

	SalesOrderFieldId                      = "id"
	SalesOrderFieldOrgId                   = "org_id"
	SalesOrderFieldOrderNumber             = "order_number"
	SalesOrderFieldSalesChannelId          = "sales_channel_id"
	SalesOrderFieldSalesPointId            = "sales_point_id"
	SalesOrderFieldCustomerReference       = "customer_reference"
	SalesOrderFieldCrmOpportunityReference = "crm_opportunity_reference"
	SalesOrderFieldCurrencyCode            = "currency_code"
	SalesOrderFieldStatus                  = "status"
	SalesOrderFieldPaymentStatus           = "payment_status"
	SalesOrderFieldFulfillmentStatus       = "fulfillment_status"
	SalesOrderFieldInvoiceStatus           = "invoice_status"
	SalesOrderFieldSubtotal                = "subtotal"
	SalesOrderFieldDiscountTotal           = "discount_total"
	SalesOrderFieldTaxTotal                = "tax_total"
	SalesOrderFieldGrandTotal              = "grand_total"
	SalesOrderFieldExchangeOfReturnId      = "exchange_of_return_id"
	SalesOrderFieldExternalReference       = "external_reference"
	SalesOrderFieldIdempotencyKey          = "idempotency_key"
	SalesOrderFieldConfirmedAt             = "confirmed_at"
	SalesOrderFieldCompletedAt             = "completed_at"
	SalesOrderFieldTaxSnapshot             = "tax_snapshot"
	SalesOrderFieldCancelledAt             = "cancelled_at"

	SalesOrderEdgeSalesChannel = "sales_channel"
	SalesOrderEdgeSalesPoint   = "sales_point"
)

//go:embed sales_order.json
var salesOrderSchemaJson string

func SalesOrderSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesOrderSchemaJson)
}

// SalesOrder is one commercial transaction: what was sold, to whom, through which channel.
//
// It carries FOUR independent status fields and they are never collapsed into one (BR §9). That is
// the single most important thing about this model. A confirmed order can be fully paid and not yet
// delivered, or delivered and unpaid, or complete but with its VAT invoice rejected by the tax
// authority — a single status would have to invent an ordering between those that the business does
// not have, and every consumer would then have to guess which dimension a given value referred to.
type SalesOrder struct {
	basemodel.DynamicModelBase
}

func NewSalesOrder() *SalesOrder {
	return &SalesOrder{basemodel.NewDynamicModel()}
}

func NewSalesOrderFrom(src dmodel.DynamicFields) *SalesOrder {
	return &SalesOrder{basemodel.NewDynamicModel(src)}
}

func (this SalesOrder) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this *SalesOrder) SetId(id *model.Id) {
	this.GetFieldData().SetModelId(basemodel.FieldId, id)
}

func (this SalesOrder) GetOrderNumber() *string {
	return this.GetFieldData().GetString(SalesOrderFieldOrderNumber)
}

func (this *SalesOrder) SetOrderNumber(number *string) {
	this.GetFieldData().SetString(SalesOrderFieldOrderNumber, number)
}

func (this SalesOrder) GetSalesChannelId() *model.Id {
	return this.GetFieldData().GetModelId(SalesOrderFieldSalesChannelId)
}

func (this *SalesOrder) SetSalesChannelId(id *model.Id) {
	this.GetFieldData().SetModelId(SalesOrderFieldSalesChannelId, id)
}

func (this SalesOrder) GetSalesPointId() *model.Id {
	return this.GetFieldData().GetModelId(SalesOrderFieldSalesPointId)
}

func (this *SalesOrder) SetSalesPointId(id *model.Id) {
	this.GetFieldData().SetModelId(SalesOrderFieldSalesPointId, id)
}

func (this SalesOrder) GetCustomerReference() *model.Id {
	return this.GetFieldData().GetModelId(SalesOrderFieldCustomerReference)
}

func (this *SalesOrder) SetCustomerReference(id *model.Id) {
	this.GetFieldData().SetModelId(SalesOrderFieldCustomerReference, id)
}

func (this SalesOrder) GetCurrencyCode() *string {
	return this.GetFieldData().GetString(SalesOrderFieldCurrencyCode)
}

func (this *SalesOrder) SetCurrencyCode(code *string) {
	this.GetFieldData().SetString(SalesOrderFieldCurrencyCode, code)
}

func (this SalesOrder) GetStatus() *string {
	return this.GetFieldData().GetString(SalesOrderFieldStatus)
}

func (this *SalesOrder) SetStatus(status *string) {
	this.GetFieldData().SetString(SalesOrderFieldStatus, status)
}

func (this SalesOrder) GetPaymentStatus() *string {
	return this.GetFieldData().GetString(SalesOrderFieldPaymentStatus)
}

func (this *SalesOrder) SetPaymentStatus(status *string) {
	this.GetFieldData().SetString(SalesOrderFieldPaymentStatus, status)
}

func (this SalesOrder) GetFulfillmentStatus() *string {
	return this.GetFieldData().GetString(SalesOrderFieldFulfillmentStatus)
}

func (this *SalesOrder) SetFulfillmentStatus(status *string) {
	this.GetFieldData().SetString(SalesOrderFieldFulfillmentStatus, status)
}

func (this SalesOrder) GetInvoiceStatus() *string {
	return this.GetFieldData().GetString(SalesOrderFieldInvoiceStatus)
}

func (this *SalesOrder) SetInvoiceStatus(status *string) {
	this.GetFieldData().SetString(SalesOrderFieldInvoiceStatus, status)
}

func (this SalesOrder) GetSubtotal() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesOrderFieldSubtotal)
}

func (this *SalesOrder) SetSubtotal(amount *decimal.Decimal) {
	this.GetFieldData().SetDecimal(SalesOrderFieldSubtotal, amount)
}

func (this SalesOrder) GetDiscountTotal() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesOrderFieldDiscountTotal)
}

func (this *SalesOrder) SetDiscountTotal(amount *decimal.Decimal) {
	this.GetFieldData().SetDecimal(SalesOrderFieldDiscountTotal, amount)
}

func (this SalesOrder) GetTaxTotal() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesOrderFieldTaxTotal)
}

func (this *SalesOrder) SetTaxTotal(amount *decimal.Decimal) {
	this.GetFieldData().SetDecimal(SalesOrderFieldTaxTotal, amount)
}

// GetTaxSnapshot returns Accounting's frozen record of how this order was taxed.
//
// Typed as `any` rather than as the accounting Snapshot struct on purpose: domain/models must not
// import another module, and a stored snapshot is read back as whatever the JSON decoder chose. A
// caller that needs the structured shape unmarshals it; a caller that only stores and returns it —
// which is most of them — passes it through untouched.
func (this SalesOrder) GetTaxSnapshot() any {
	return this.GetFieldData().GetAny(SalesOrderFieldTaxSnapshot)
}

func (this *SalesOrder) SetTaxSnapshot(snapshot any) {
	this.GetFieldData().SetAny(SalesOrderFieldTaxSnapshot, snapshot)
}

func (this SalesOrder) GetGrandTotal() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesOrderFieldGrandTotal)
}

func (this *SalesOrder) SetGrandTotal(amount *decimal.Decimal) {
	this.GetFieldData().SetDecimal(SalesOrderFieldGrandTotal, amount)
}

func (this SalesOrder) GetExternalReference() *string {
	return this.GetFieldData().GetString(SalesOrderFieldExternalReference)
}

func (this *SalesOrder) SetExternalReference(ref *string) {
	this.GetFieldData().SetString(SalesOrderFieldExternalReference, ref)
}

func (this SalesOrder) GetIdempotencyKey() *string {
	return this.GetFieldData().GetString(SalesOrderFieldIdempotencyKey)
}

func (this *SalesOrder) SetIdempotencyKey(key *string) {
	this.GetFieldData().SetString(SalesOrderFieldIdempotencyKey, key)
}

func (this SalesOrder) GetConfirmedAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(SalesOrderFieldConfirmedAt)
}

func (this *SalesOrder) SetConfirmedAt(at *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(SalesOrderFieldConfirmedAt, at)
}

func (this SalesOrder) GetCompletedAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(SalesOrderFieldCompletedAt)
}

func (this *SalesOrder) SetCompletedAt(at *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(SalesOrderFieldCompletedAt, at)
}

func (this SalesOrder) GetCancelledAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(SalesOrderFieldCancelledAt)
}

func (this *SalesOrder) SetCancelledAt(at *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(SalesOrderFieldCancelledAt, at)
}

func (this SalesOrder) GetIsArchived() *bool {
	return this.GetFieldData().GetBool(basemodel.FieldIsArchived)
}

// IsConfirmed reports whether the sale has been committed to.
//
// This is the gate on the snapshot fields: before confirmation the document is a draft whose lines
// may be repriced freely, after it the numbers are what the business promised the customer and are
// immutable (BR §11). Every rule that edits a line has to ask this first.
func (this SalesOrder) IsConfirmed() bool {
	status := this.GetStatus()
	if status == nil {
		return false
	}
	switch SalesOrderStatus(*status) {
	case SalesOrderStatusConfirmed, SalesOrderStatusCompleted:
		return true
	}
	return false
}

// IsEditable reports whether lines may still be added, changed or removed.
//
// Only a draft qualifies. A cancelled order is deliberately NOT editable even though it will never
// be delivered: it is a record of something that was attempted, and rewriting it would destroy the
// evidence of what was attempted.
func (this SalesOrder) IsEditable() bool {
	status := this.GetStatus()
	if status == nil {
		return false
	}
	if SalesOrderStatus(*status) != SalesOrderStatusDraft {
		return false
	}
	archived := this.GetIsArchived()
	return archived == nil || !*archived
}

// IsTerminal reports whether the order has reached a state it will never leave.
//
// Completed and cancelled are both terminal for the ORDER status specifically. The other three
// dimensions can still move afterwards — a completed order can be refunded, returned, and have its
// invoice issued or rejected long after the sale itself is over, which is precisely why the four
// statuses are separate fields.
func (this SalesOrder) IsTerminal() bool {
	status := this.GetStatus()
	if status == nil {
		return false
	}
	switch SalesOrderStatus(*status) {
	case SalesOrderStatusCompleted, SalesOrderStatusCancelled:
		return true
	}
	return false
}

// GetExchangeOfReturnId returns the return this order replaces, or nil.
//
// AN EXCHANGE IS A RETURN PLUS A NEW SALE (BR 87.5), joined by this column and never by editing the
// product variant on a historical order. The original sale really did sell what it sold; rewriting it
// would restate a transaction the customer already paid for, invalidate any fiscal document issued
// against it, and leave the returned goods unaccounted for.
func (this SalesOrder) GetExchangeOfReturnId() *model.Id {
	return this.GetFieldData().GetModelId(SalesOrderFieldExchangeOfReturnId)
}

// IsExchange reports whether this order is the outgoing half of an exchange.
//
// The question a fulfilment or refund decision turns on: an exchange's new order is paid for by the
// return it replaces rather than by fresh money, so treating it as an ordinary sale would either
// charge the customer twice or leave the return unrefunded.
func (this SalesOrder) IsExchange() bool {
	return this.GetExchangeOfReturnId() != nil
}

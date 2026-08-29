package models

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// The billing trio: a settlement unit, its allocations, and the lineage between bills.
//
// A bill is not a VAT invoice: one sale may need several settlement units and one legal document,
// or the reverse. Nothing here should ever grow an invoice number.

const (
	SalesBillSchemaName = "sales_bill"

	SalesBillFieldId            = basemodel.FieldId
	SalesBillFieldOrgId         = basemodel.FieldOrgId
	SalesBillFieldBillNumber    = "bill_number"
	SalesBillFieldSalesOrderId  = "sales_order_id"
	SalesBillFieldStatus        = "status"
	SalesBillFieldPaymentStatus = "payment_status"
	SalesBillFieldCurrencyCode  = "currency_code"
	SalesBillFieldSubtotal      = "subtotal"
	SalesBillFieldDiscountTotal = "discount_total"
	SalesBillFieldTaxTotal      = "tax_total"
	SalesBillFieldTotalAmount   = "total_amount"
	SalesBillFieldSettledAt     = "settled_at"
	SalesBillFieldCancelledAt   = "cancelled_at"
	SalesBillFieldIsArchived    = basemodel.FieldIsArchived

	SalesBillEdgeSalesOrder = "sales_order"
)

const (
	SalesBillLineSchemaName = "sales_bill_line"

	SalesBillLineFieldId                   = basemodel.FieldId
	SalesBillLineFieldOrgId                = basemodel.FieldOrgId
	SalesBillLineFieldSalesBillId          = "sales_bill_id"
	SalesBillLineFieldSalesOrderLineId     = "sales_order_line_id"
	SalesBillLineFieldQuantity             = "quantity"
	SalesBillLineFieldAllocatedNetAmount   = "allocated_net_amount"
	SalesBillLineFieldAllocatedTaxAmount   = "allocated_tax_amount"
	SalesBillLineFieldAllocatedTotalAmount = "allocated_total_amount"

	SalesBillLineEdgeSalesBill      = "sales_bill"
	SalesBillLineEdgeSalesOrderLine = "sales_order_line"
)

const (
	SalesBillRelationSchemaName = "sales_bill_relation"

	SalesBillRelationFieldId           = basemodel.FieldId
	SalesBillRelationFieldOrgId        = basemodel.FieldOrgId
	SalesBillRelationFieldSourceBillId = "source_bill_id"
	SalesBillRelationFieldTargetBillId = "target_bill_id"
	SalesBillRelationFieldType         = "relation_type"

	SalesBillRelationEdgeSourceBill = "source_bill"
	SalesBillRelationEdgeTargetBill = "target_bill"
)

//go:embed sales_bill.json
var salesBillSchemaJson string

func SalesBillSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesBillSchemaJson)
}

//go:embed sales_bill_line.json
var salesBillLineSchemaJson string

func SalesBillLineSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesBillLineSchemaJson)
}

//go:embed sales_bill_relation.json
var salesBillRelationSchemaJson string

func SalesBillRelationSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesBillRelationSchemaJson)
}

// SalesBill is one settlement unit of a sale.
type SalesBill struct {
	basemodel.DynamicModelBase
}

func NewSalesBill() *SalesBill {
	return &SalesBill{basemodel.NewDynamicModel()}
}

func NewSalesBillFrom(src dmodel.DynamicFields) *SalesBill {
	return &SalesBill{basemodel.NewDynamicModel(src)}
}

func (this SalesBill) GetId() *model.Id {
	return this.GetFieldData().GetModelId(SalesBillFieldId)
}

func (this SalesBill) GetBillNumber() *string {
	return this.GetFieldData().GetString(SalesBillFieldBillNumber)
}

func (this SalesBill) GetSalesOrderId() *model.Id {
	return this.GetFieldData().GetModelId(SalesBillFieldSalesOrderId)
}

func (this SalesBill) GetStatus() *string {
	return this.GetFieldData().GetString(SalesBillFieldStatus)
}

func (this *SalesBill) SetStatus(status *string) {
	this.GetFieldData().SetString(SalesBillFieldStatus, status)
}

func (this SalesBill) GetPaymentStatus() *string {
	return this.GetFieldData().GetString(SalesBillFieldPaymentStatus)
}

func (this *SalesBill) SetPaymentStatus(status *string) {
	this.GetFieldData().SetString(SalesBillFieldPaymentStatus, status)
}

func (this SalesBill) GetCurrencyCode() *string {
	return this.GetFieldData().GetString(SalesBillFieldCurrencyCode)
}

func (this SalesBill) GetTotalAmount() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesBillFieldTotalAmount)
}

func (this *SalesBill) SetTotalAmount(amount *decimal.Decimal) {
	this.GetFieldData().SetDecimal(SalesBillFieldTotalAmount, amount)
}

func (this SalesBill) GetSettledAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(SalesBillFieldSettledAt)
}

func (this *SalesBill) SetSettledAt(at *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(SalesBillFieldSettledAt, at)
}

func (this SalesBill) GetCancelledAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(SalesBillFieldCancelledAt)
}

func (this *SalesBill) SetCancelledAt(at *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(SalesBillFieldCancelledAt, at)
}

// IsOpen reports whether this bill may still be split, merged or paid into.
func (this SalesBill) IsOpen() bool {
	status := this.GetStatus()
	return status != nil && *status == string(SalesBillStatusOpen)
}

// IsSettled reports whether the money is fully in and the allocations have frozen.
func (this SalesBill) IsSettled() bool {
	status := this.GetStatus()
	return status != nil && *status == string(SalesBillStatusSettled)
}

// SalesBillLine is one order line's contribution to one bill.
type SalesBillLine struct {
	basemodel.DynamicModelBase
}

func NewSalesBillLine() *SalesBillLine {
	return &SalesBillLine{basemodel.NewDynamicModel()}
}

func NewSalesBillLineFrom(src dmodel.DynamicFields) *SalesBillLine {
	return &SalesBillLine{basemodel.NewDynamicModel(src)}
}

func (this SalesBillLine) GetId() *model.Id {
	return this.GetFieldData().GetModelId(SalesBillLineFieldId)
}

func (this SalesBillLine) GetSalesBillId() *model.Id {
	return this.GetFieldData().GetModelId(SalesBillLineFieldSalesBillId)
}

func (this SalesBillLine) GetSalesOrderLineId() *model.Id {
	return this.GetFieldData().GetModelId(SalesBillLineFieldSalesOrderLineId)
}

func (this SalesBillLine) GetQuantity() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesBillLineFieldQuantity)
}

func (this SalesBillLine) GetAllocatedTotalAmount() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesBillLineFieldAllocatedTotalAmount)
}

// SumAllocatedTotal adds up what a set of allocations comes to. One implementation, so callers
// cannot disagree about whether an absent value counts as zero.
func SumAllocatedTotal(allocations []dmodel.DynamicFields) decimal.Decimal {
	total := decimal.Zero
	for _, allocation := range allocations {
		total = total.Add(allocatedTotalOf(allocation))
	}
	return total
}

// allocatedTotalOf reads one allocation's total in whatever shape it arrived in. Not GetDecimal:
// that accessor panics on anything but decimal.Decimal, and a jsonb round trip returns a decimal as
// a string. An unreadable amount counts as zero, which makes the bills look short and the mutation
// gets refused - safer than crediting an amount that may not be what was stored.
func allocatedTotalOf(allocation dmodel.DynamicFields) decimal.Decimal {
	value, present := allocation[SalesBillLineFieldAllocatedTotalAmount]
	if !present || value == nil {
		return decimal.Zero
	}
	switch typed := value.(type) {
	case decimal.Decimal:
		return typed
	case *decimal.Decimal:
		if typed != nil {
			return *typed
		}
	case string:
		if parsed, err := decimal.NewFromString(typed); err == nil {
			return parsed
		}
	case float64:
		return decimal.NewFromFloat(typed)
	}
	return decimal.Zero
}

// SalesBillRelation records that one bill became another.
type SalesBillRelation struct {
	basemodel.DynamicModelBase
}

func NewSalesBillRelation() *SalesBillRelation {
	return &SalesBillRelation{basemodel.NewDynamicModel()}
}

func NewSalesBillRelationFrom(src dmodel.DynamicFields) *SalesBillRelation {
	return &SalesBillRelation{basemodel.NewDynamicModel(src)}
}

func (this SalesBillRelation) GetSourceBillId() *model.Id {
	return this.GetFieldData().GetModelId(SalesBillRelationFieldSourceBillId)
}

func (this SalesBillRelation) GetTargetBillId() *model.Id {
	return this.GetFieldData().GetModelId(SalesBillRelationFieldTargetBillId)
}

func (this SalesBillRelation) GetRelationType() *string {
	return this.GetFieldData().GetString(SalesBillRelationFieldType)
}

const (
	SalesPaymentSchemaName = "sales_payment"

	SalesPaymentFieldId                    = basemodel.FieldId
	SalesPaymentFieldOrgId                 = basemodel.FieldOrgId
	SalesPaymentFieldSalesBillId           = "sales_bill_id"
	SalesPaymentFieldPaymentMethodId       = "payment_method_id"
	SalesPaymentFieldMethodCodeSnapshot    = "payment_method_code_snapshot"
	SalesPaymentFieldAmount                = "amount"
	SalesPaymentFieldCurrencyCode          = "currency_code"
	SalesPaymentFieldStatus                = "status"
	SalesPaymentFieldExternalTransactionId = "external_transaction_id"
	SalesPaymentFieldProviderReference     = "provider_reference"
	SalesPaymentFieldPaidAt                = "paid_at"

	SalesPaymentEdgeSalesBill = "sales_bill"
)

//go:embed sales_payment.json
var salesPaymentSchemaJson string

func SalesPaymentSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesPaymentSchemaJson)
}

// SalesPayment is one movement of money against one bill.
type SalesPayment struct {
	basemodel.DynamicModelBase
}

func NewSalesPayment() *SalesPayment {
	return &SalesPayment{basemodel.NewDynamicModel()}
}

func NewSalesPaymentFrom(src dmodel.DynamicFields) *SalesPayment {
	return &SalesPayment{basemodel.NewDynamicModel(src)}
}

func (this SalesPayment) GetId() *model.Id {
	return this.GetFieldData().GetModelId(SalesPaymentFieldId)
}

func (this SalesPayment) GetSalesBillId() *model.Id {
	return this.GetFieldData().GetModelId(SalesPaymentFieldSalesBillId)
}

func (this SalesPayment) GetPaymentMethodId() *model.Id {
	return this.GetFieldData().GetModelId(SalesPaymentFieldPaymentMethodId)
}

func (this SalesPayment) GetAmount() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesPaymentFieldAmount)
}

func (this SalesPayment) GetStatus() *string {
	return this.GetFieldData().GetString(SalesPaymentFieldStatus)
}

func (this *SalesPayment) SetStatus(status *string) {
	this.GetFieldData().SetString(SalesPaymentFieldStatus, status)
}

func (this SalesPayment) GetExternalTransactionId() *string {
	return this.GetFieldData().GetString(SalesPaymentFieldExternalTransactionId)
}

// IsCaptured reports whether this payment's money is actually in. An authorization does not count:
// it is a hold the provider may still release, and counting it would settle a bill against funds
// that never arrived.
func (this SalesPayment) IsCaptured() bool {
	status := this.GetStatus()
	return status != nil && *status == string(SalesPaymentStatusCaptured)
}

// SumCapturedAmount adds up the money actually taken against a bill. Only captured payments count,
// so that "how much has been paid" has exactly one answer.
func SumCapturedAmount(payments []dmodel.DynamicFields) decimal.Decimal {
	total := decimal.Zero
	for _, payment := range payments {
		if !NewSalesPaymentFrom(payment).IsCaptured() {
			continue
		}
		if amount := NewSalesPaymentFrom(payment).GetAmount(); amount != nil {
			total = total.Add(*amount)
		}
	}
	return total
}

const (
	SalesFulfillmentRequestSchemaName = "sales_fulfillment_request"

	SalesFulfillmentRequestFieldId           = basemodel.FieldId
	SalesFulfillmentRequestFieldOrgId        = basemodel.FieldOrgId
	SalesFulfillmentRequestFieldSalesOrderId = "sales_order_id"
	SalesFulfillmentRequestFieldRequestType  = "request_type"
	SalesFulfillmentRequestFieldStatus       = "status"
	SalesFulfillmentRequestFieldInventoryRef = "inventory_reference"
	SalesFulfillmentRequestFieldFailReason   = "failure_reason"
	SalesFulfillmentRequestFieldRequestedAt  = "requested_at"
	SalesFulfillmentRequestFieldCompletedAt  = "completed_at"

	SalesFulfillmentRequestEdgeSalesOrder = "sales_order"
)

const (
	SalesFulfillmentRequestLineSchemaName = "sales_fulfillment_request_line"

	SalesFulfillmentLineFieldId          = basemodel.FieldId
	SalesFulfillmentLineFieldOrgId       = basemodel.FieldOrgId
	SalesFulfillmentLineFieldRequestId   = "sales_fulfillment_request_id"
	SalesFulfillmentLineFieldOrderLineId = "sales_order_line_id"
	SalesFulfillmentLineFieldQuantity    = "quantity"

	SalesFulfillmentLineEdgeRequest   = "sales_fulfillment_request"
	SalesFulfillmentLineEdgeOrderLine = "sales_order_line"
)

//go:embed sales_fulfillment_request.json
var salesFulfillmentRequestSchemaJson string

func SalesFulfillmentRequestSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesFulfillmentRequestSchemaJson)
}

//go:embed sales_fulfillment_request_line.json
var salesFulfillmentRequestLineSchemaJson string

func SalesFulfillmentRequestLineSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesFulfillmentRequestLineSchemaJson)
}

// SalesFulfillmentRequest is one thing Sales asked Inventory to do.
type SalesFulfillmentRequest struct {
	basemodel.DynamicModelBase
}

func NewSalesFulfillmentRequestFrom(src dmodel.DynamicFields) *SalesFulfillmentRequest {
	return &SalesFulfillmentRequest{basemodel.NewDynamicModel(src)}
}

func (this SalesFulfillmentRequest) GetStatus() *string {
	return this.GetFieldData().GetString(SalesFulfillmentRequestFieldStatus)
}

func (this SalesFulfillmentRequest) GetRequestType() *string {
	return this.GetFieldData().GetString(SalesFulfillmentRequestFieldRequestType)
}

// IsCompleted reports whether the goods actually moved. An ACCEPTED request has stock held but
// nothing moved, so counting it would tell a customer their goods had shipped when they had not.
func (this SalesFulfillmentRequest) IsCompleted() bool {
	status := this.GetStatus()
	return status != nil && *status == string(SalesFulfillmentStatusCompleted)
}

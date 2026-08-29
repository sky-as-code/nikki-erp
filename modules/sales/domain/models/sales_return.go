package models

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// The return trio: a return document, the lines it covers, and the refunds that settle it.
//
// A return is a document, never an edit: a partial return references the original order rather than
// rewriting it, because editing would restate a transaction already paid for and reported. It
// carries three status columns because its three consequences fail independently — goods back, money
// back, tax authority told. The return is commercially complete once the first two are done; the
// fiscal status never gates the others.

const (
	SalesReturnSchemaName = "sales_return"

	SalesReturnFieldId                     = basemodel.FieldId
	SalesReturnFieldOrgId                  = basemodel.FieldOrgId
	SalesReturnFieldReturnNumber           = "return_number"
	SalesReturnFieldSalesOrderId           = "sales_order_id"
	SalesReturnFieldStatus                 = "status"
	SalesReturnFieldInventoryReturnStatus  = "inventory_return_status"
	SalesReturnFieldRefundStatus           = "refund_status"
	SalesReturnFieldFiscalAdjustmentStatus = "fiscal_adjustment_status"
	SalesReturnFieldReason                 = "reason"
	SalesReturnFieldInventoryDisposition   = "inventory_disposition"
	SalesReturnFieldRefundTotal            = "refund_total"
	SalesReturnFieldInventoryReference     = "inventory_reference"
	SalesReturnFieldFailureReason          = "failure_reason"
	SalesReturnFieldRequestedAt            = "requested_at"
	SalesReturnFieldCompletedAt            = "completed_at"
	SalesReturnFieldCancelledAt            = "cancelled_at"
	SalesReturnFieldCreatedAt              = basemodel.FieldCreatedAt
	SalesReturnFieldUpdatedAt              = basemodel.FieldUpdatedAt
	SalesReturnFieldEtag                   = basemodel.FieldEtag

	SalesReturnEdgeSalesOrder = "sales_order"
)

const (
	SalesReturnLineSchemaName = "sales_return_line"

	SalesReturnLineFieldId                      = basemodel.FieldId
	SalesReturnLineFieldOrgId                   = basemodel.FieldOrgId
	SalesReturnLineFieldSalesReturnId           = "sales_return_id"
	SalesReturnLineFieldSalesOrderLineId        = "sales_order_line_id"
	SalesReturnLineFieldQuantity                = "quantity"
	SalesReturnLineFieldRefundAmount            = "refund_amount"
	SalesReturnLineFieldRefundTaxAmount         = "refund_tax_amount"
	SalesReturnLineFieldRequiresInventoryReturn = "requires_inventory_return"

	SalesReturnLineEdgeSalesReturn    = "sales_return"
	SalesReturnLineEdgeSalesOrderLine = "sales_order_line"
)

const (
	SalesRefundPaymentSchemaName = "sales_refund_payment"

	SalesRefundPaymentFieldId                     = basemodel.FieldId
	SalesRefundPaymentFieldOrgId                  = basemodel.FieldOrgId
	SalesRefundPaymentFieldSalesReturnId          = "sales_return_id"
	SalesRefundPaymentFieldOriginalSalesPaymentId = "original_sales_payment_id"
	SalesRefundPaymentFieldAmount                 = "amount"
	SalesRefundPaymentFieldCurrencyCode           = "currency_code"
	SalesRefundPaymentFieldStatus                 = "status"
	SalesRefundPaymentFieldProviderReference      = "provider_reference"
	SalesRefundPaymentFieldFailureReason          = "failure_reason"
	SalesRefundPaymentFieldCompletedAt            = "completed_at"

	SalesRefundPaymentEdgeSalesReturn     = "sales_return"
	SalesRefundPaymentEdgeOriginalPayment = "original_payment"
)

//go:embed sales_return.json
var salesReturnSchemaJson string

func SalesReturnSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesReturnSchemaJson)
}

//go:embed sales_return_line.json
var salesReturnLineSchemaJson string

func SalesReturnLineSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesReturnLineSchemaJson)
}

//go:embed sales_refund_payment.json
var salesRefundPaymentSchemaJson string

func SalesRefundPaymentSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesRefundPaymentSchemaJson)
}

// SalesReturn is one return document against an order.
type SalesReturn struct {
	basemodel.DynamicModelBase
}

func NewSalesReturn() *SalesReturn {
	return &SalesReturn{basemodel.NewDynamicModel()}
}

func NewSalesReturnFrom(src dmodel.DynamicFields) *SalesReturn {
	return &SalesReturn{basemodel.NewDynamicModel(src)}
}

func (this SalesReturn) GetId() *model.Id {
	return this.GetFieldData().GetModelId(SalesReturnFieldId)
}

func (this SalesReturn) GetSalesOrderId() *model.Id {
	return this.GetFieldData().GetModelId(SalesReturnFieldSalesOrderId)
}

func (this SalesReturn) GetStatus() *string {
	return this.GetFieldData().GetString(SalesReturnFieldStatus)
}

func (this SalesReturn) GetInventoryReturnStatus() *string {
	return this.GetFieldData().GetString(SalesReturnFieldInventoryReturnStatus)
}

func (this SalesReturn) GetRefundStatus() *string {
	return this.GetFieldData().GetString(SalesReturnFieldRefundStatus)
}

func (this SalesReturn) GetFiscalAdjustmentStatus() *string {
	return this.GetFieldData().GetString(SalesReturnFieldFiscalAdjustmentStatus)
}

func (this SalesReturn) GetRefundTotal() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesReturnFieldRefundTotal)
}

// SalesReturnLine is one order line coming back, in whole or in part.
type SalesReturnLine struct {
	basemodel.DynamicModelBase
}

func NewSalesReturnLine() *SalesReturnLine {
	return &SalesReturnLine{basemodel.NewDynamicModel()}
}

func NewSalesReturnLineFrom(src dmodel.DynamicFields) *SalesReturnLine {
	return &SalesReturnLine{basemodel.NewDynamicModel(src)}
}

func (this SalesReturnLine) GetId() *model.Id {
	return this.GetFieldData().GetModelId(SalesReturnLineFieldId)
}

func (this SalesReturnLine) GetSalesOrderLineId() *model.Id {
	return this.GetFieldData().GetModelId(SalesReturnLineFieldSalesOrderLineId)
}

func (this SalesReturnLine) GetQuantity() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesReturnLineFieldQuantity)
}

func (this SalesReturnLine) GetRefundAmount() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesReturnLineFieldRefundAmount)
}

// SalesRefundPayment is one leg of a refund, against one original payment.
type SalesRefundPayment struct {
	basemodel.DynamicModelBase
}

func NewSalesRefundPayment() *SalesRefundPayment {
	return &SalesRefundPayment{basemodel.NewDynamicModel()}
}

func NewSalesRefundPaymentFrom(src dmodel.DynamicFields) *SalesRefundPayment {
	return &SalesRefundPayment{basemodel.NewDynamicModel(src)}
}

func (this SalesRefundPayment) GetId() *model.Id {
	return this.GetFieldData().GetModelId(SalesRefundPaymentFieldId)
}

func (this SalesRefundPayment) GetOriginalSalesPaymentId() *model.Id {
	return this.GetFieldData().GetModelId(SalesRefundPaymentFieldOriginalSalesPaymentId)
}

func (this SalesRefundPayment) GetAmount() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesRefundPaymentFieldAmount)
}

func (this SalesRefundPayment) GetStatus() *string {
	return this.GetFieldData().GetString(SalesRefundPaymentFieldStatus)
}

// SumCompletedRefunds totals only the refund legs that actually went back. A pending refund is a
// promise the provider has not honoured, and counting it would report a customer repaid who is
// still waiting.
func SumCompletedRefunds(refunds []dmodel.DynamicFields) decimal.Decimal {
	total := decimal.Zero
	for _, record := range refunds {
		refund := NewSalesRefundPaymentFrom(record)
		status := refund.GetStatus()
		if status == nil || *status != string(SalesRefundPaymentStatusCompleted) {
			continue
		}
		if amount := refund.GetAmount(); amount != nil {
			total = total.Add(*amount)
		}
	}
	return total
}

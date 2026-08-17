package models

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	TransactionSchemaName = "paymentinvoice_transaction"

	TransactionFieldId              = basemodel.FieldId
	TransactionFieldOrderId         = "order_id"
	TransactionFieldOrderBusinessId = "order_business_id"
	TransactionFieldStatus          = "status"
	TransactionFieldAmount          = "amount"
	TransactionFieldCurrencyId      = "currency_id"
	TransactionFieldPaymentMethodId = "payment_method_id"
	TransactionFieldTransactionType = "transaction_type"
	TransactionFieldContent         = "content"

	TransactionFieldRefTransactionId = "ref_transaction_id"
	TransactionFieldRefPayload       = "ref_payload"
	TransactionFieldOrgId            = "org_id"
)

const (
	TransactionStatusPending   = "pending"
	TransactionStatusCompleted = "completed"
	TransactionStatusFailed    = "failed"
	TransactionStatusCanceled  = "canceled"
)

const (
	TransactionTypePayment = "payment"
	TransactionTypeRefund  = "refund"
)

//go:embed transaction.json
var transactionSchemaJson string

func TransactionSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(transactionSchemaJson)
}

// Transaction is one payment or refund attempt recorded against an order.
// Table: paymentinvoice_transactions.
//
// An order has one payment transaction and one further transaction per refund, which is what
// makes the money actually moved reconstructable from the transactions alone.
//
// order_business_id duplicates the order's quoted identifier deliberately: reconciling against a
// gateway's statement is a lookup by that value, and it should not need a join.
//
// Rows here are written by the payment flow and the gateway callbacks, never by a person — a row
// someone could edit would be worthless as the evidence it exists to be.
type Transaction struct {
	basemodel.DynamicModelBase
}

func NewTransaction() *Transaction {
	return &Transaction{basemodel.NewDynamicModel()}
}

func NewTransactionFrom(src dmodel.DynamicFields) *Transaction {
	return &Transaction{basemodel.NewDynamicModel(src)}
}

func (this Transaction) GetOrderId() *model.Id {
	return this.GetFieldData().GetModelId(TransactionFieldOrderId)
}

func (this *Transaction) SetOrderId(v *model.Id) {
	this.GetFieldData().SetModelId(TransactionFieldOrderId, v)
}

func (this Transaction) GetOrderBusinessId() *string {
	return this.GetFieldData().GetString(TransactionFieldOrderBusinessId)
}

func (this *Transaction) SetOrderBusinessId(v *string) {
	this.GetFieldData().SetString(TransactionFieldOrderBusinessId, v)
}

func (this Transaction) GetStatus() *string {
	return this.GetFieldData().GetString(TransactionFieldStatus)
}

func (this *Transaction) SetStatus(v *string) {
	this.GetFieldData().SetString(TransactionFieldStatus, v)
}

func (this Transaction) GetAmount() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(TransactionFieldAmount)
}

func (this *Transaction) SetAmount(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(TransactionFieldAmount, v)
}

func (this Transaction) GetCurrencyId() *model.Id {
	return this.GetFieldData().GetModelId(TransactionFieldCurrencyId)
}

func (this *Transaction) SetCurrencyId(v *model.Id) {
	this.GetFieldData().SetModelId(TransactionFieldCurrencyId, v)
}

func (this Transaction) GetPaymentMethodId() *model.Id {
	return this.GetFieldData().GetModelId(TransactionFieldPaymentMethodId)
}

func (this *Transaction) SetPaymentMethodId(v *model.Id) {
	this.GetFieldData().SetModelId(TransactionFieldPaymentMethodId, v)
}

func (this Transaction) GetTransactionType() *string {
	return this.GetFieldData().GetString(TransactionFieldTransactionType)
}

func (this *Transaction) SetTransactionType(v *string) {
	this.GetFieldData().SetString(TransactionFieldTransactionType, v)
}

func (this Transaction) GetContent() *string {
	return this.GetFieldData().GetString(TransactionFieldContent)
}

func (this *Transaction) SetContent(v *string) {
	this.GetFieldData().SetString(TransactionFieldContent, v)
}

func (this Transaction) GetRefTransactionId() *string {
	return this.GetFieldData().GetString(TransactionFieldRefTransactionId)
}

func (this *Transaction) SetRefTransactionId(v *string) {
	this.GetFieldData().SetString(TransactionFieldRefTransactionId, v)
}

func (this Transaction) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(TransactionFieldOrgId)
}

func (this *Transaction) SetOrgId(v *model.Id) {
	this.GetFieldData().SetModelId(TransactionFieldOrgId, v)
}

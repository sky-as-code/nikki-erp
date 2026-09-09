// Package order is the Payment & Invoice module's port for taking money and giving it back.
//
// It exists because these two operations are the only ones another module has any business
// calling. Everything else this module does — recording invoices, listing transactions, managing
// payment methods and profiles — is CRUD served by the resource engine over REST, and a caller
// that wants it should use that surface rather than reach into the domain.
//
// The shapes here are deliberately the ones the standalone NestJS service exposed over HTTP,
// because the callers being migrated off that service have to keep working unchanged: an amount,
// a payment method, some free text in; an order identifier and something for the payer to pay
// with, out.
package order

import (
	"github.com/shopspring/decimal"

	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
)

// CreatePaymentCommand opens an order and asks its gateway to start collecting.
type CreatePaymentCommand struct {
	// OrgId is the organization the order belongs to. Required: it is a column on every record
	// this module writes, and it cannot be derived from the request context — a caller may belong
	// to several organizations, and guessing which one owns the money would be worse than asking.
	OrgId string `json:"org_id"`

	Amount decimal.Decimal `json:"amount"`

	// PaymentMethodId and PaymentMethodCode are two ways of naming the same row, and exactly one
	// is needed. The id is what the REST surface uses, because that is what a picker holds; the
	// code is "momo", "vietqr", "mpos" — what the service this module supersedes was called with,
	// and what a caller migrating off it already has. The id wins when both are given.
	PaymentMethodId   string `json:"payment_method_id,omitempty"`
	PaymentMethodCode string `json:"payment_method_code,omitempty"`

	// PaymentProfileId names the merchant account to collect into. Optional: an order without one
	// is collected with the credentials in this deployment's configuration.
	PaymentProfileId string `json:"payment_profile_id,omitempty"`

	// Source names the system the order came from, and becomes part of the order identifier.
	// Empty defaults to the vending machines, which were the only caller of the old service.
	Source string `json:"source,omitempty"`

	// Content is what the payer sees on their statement. Empty defaults to the order identifier,
	// so a transfer can still be traced back.
	Content *string `json:"content,omitempty"`

	// ReturnUrl is where the ordering system wants to be notified once the payment settles.
	ReturnUrl *string `json:"return_url,omitempty"`

	// Metadata is the method-specific input, uninterpreted. Only the selected gateway adapter
	// reads it — a card terminal needs pos_id here, a wallet needs nothing.
	Metadata map[string]any `json:"metadata,omitempty"`
}

// CreatePaymentResultData is what the payer needs in order to pay.
type CreatePaymentResultData struct {
	// OrderId is the identifier the ordering system quotes, and the one a refund is filed under.
	OrderId string `json:"order_id"`

	// OrderCode is the identifier the gateway knows the order by, and the key its callback
	// arrives under. Kept distinct from OrderId: they are different values with different
	// audiences, and a caller that stores one where the other belongs cannot refund the order.
	OrderCode string `json:"order_code"`

	// QrCodeUrl and PayUrl are both empty for a card terminal, where the prompt is pushed to the
	// device the customer is standing at and there is nothing to show on a screen.
	QrCodeUrl string `json:"qr_code_url"`
	PayUrl    string `json:"pay_url"`
}

type CreatePaymentResult = dyn.OpResult[CreatePaymentResultData]

// RefundCommand gives money back against an order the caller quotes.
type RefundCommand struct {
	// OrderId and OrderCode identify the same order two ways, and exactly one is needed. Both are
	// accepted because the two identifiers were handed out together by the service this module
	// supersedes, and callers migrating off it kept whichever one they happened to store.
	// OrderId wins when both are given.
	OrderId   string `json:"order_id,omitempty"`
	OrderCode string `json:"order_code,omitempty"`

	Amount  decimal.Decimal `json:"amount"`
	Content *string         `json:"content,omitempty"`
}

// RefundResultData reports what was returned and what remains.
type RefundResultData struct {
	OrderId      string          `json:"order_id"`
	RefundAmount decimal.Decimal `json:"refund_amount"`

	// RestedAmount is what is left of the order after this refund. The spelling is the old
	// service's and is kept: the ordering system reads this key.
	RestedAmount decimal.Decimal `json:"rested_amount"`
}

type RefundResult = dyn.OpResult[RefundResultData]

// GetOrderStatusQuery names the order to read back. OrderId and OrderCode identify it the same two
// ways a refund does, and exactly one is needed; OrderId wins when both are given.
type GetOrderStatusQuery struct {
	OrderId   string `json:"order_id,omitempty"`
	OrderCode string `json:"order_code,omitempty"`
}

// GetOrderStatusResultData is what a caller reconciling a payment needs to decide what to do.
type GetOrderStatusResultData struct {
	OrderId   string `json:"order_id"`
	OrderCode string `json:"order_code"`

	// Status is the order's own status — pending, processing, payment_success and the rest. A caller
	// maps it onto its own vocabulary rather than this module inventing a shared one, because what
	// "settled" means to a till is not what it means to a ledger.
	Status string `json:"status"`

	Amount       decimal.Decimal `json:"amount"`
	RefundAmount decimal.Decimal `json:"refund_amount"`

	// RefTransactionId is the gateway's identifier for the payment, once there is one. Empty while
	// the order is still pending: it is issued when money actually moves, which is exactly why it
	// cannot be the key a caller correlates on beforehand.
	RefTransactionId string `json:"ref_transaction_id,omitempty"`
}

type GetOrderStatusResult = dyn.OpResult[GetOrderStatusResultData]

// The order statuses a caller reading GetOrderStatusResultData.Status may see.
//
// They are re-exported here, beside the field that carries them, because the values are part of
// this port's contract: a caller has to branch on them, and reaching into this module's domain
// models to name them would couple it to a package it has no other business with. The definitions
// stay in domain/models — these are aliases of the same constants, not a second set.
const (
	OrderStatusPending        = models.OrderStatusPending
	OrderStatusProcessing     = models.OrderStatusProcessing
	OrderStatusPaymentSuccess = models.OrderStatusPaymentSuccess
	OrderStatusPaymentFailed  = models.OrderStatusPaymentFailed
	OrderStatusCanceled       = models.OrderStatusCanceled
	OrderStatusRefundSuccess  = models.OrderStatusRefundSuccess
	OrderStatusRefundFailed   = models.OrderStatusRefundFailed
	OrderStatusExpired        = models.OrderStatusExpired
)

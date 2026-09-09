package external

import (
	"github.com/shopspring/decimal"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// PaymentOrderExtService is Sales' port onto paymentinvoice's gateway, for the tenders where money
// has to be collected rather than simply counted.
//
// It is separate from PaymentMethodExtService on purpose. That port reads master data and judges
// what may be offered; this one moves money. Binding them together would let any code holding a
// method list also start a collection.
//
// Sales never learns which gateway serves a method — momo, vietqr and mpos are paymentinvoice's
// business. Sales asks for an order and is handed something for the customer to pay with.
type PaymentOrderExtService interface {
	// CreatePayment opens an order with the provider and returns what the payer needs.
	//
	// The order is opened, not settled: the customer has not paid when this returns. The verdict
	// arrives later through the settlement event, and the reconciliation sweep asks GetOrderStatus
	// for the ones that never arrive.
	CreatePayment(
		ctx corectx.Context, cmd CreateGatewayPaymentCommand,
	) (*CreateGatewayPaymentResult, error)

	// Refund gives money back against an order. Synchronous, unlike collection: the provider answers
	// on the call, and paymentinvoice holds the guard that stops an order being over-refunded.
	Refund(ctx corectx.Context, cmd RefundGatewayPaymentCommand) (*RefundGatewayPaymentResult, error)

	// GetOrderStatus reads back where an order stands, for reconciling a payment whose settlement
	// never arrived.
	GetOrderStatus(ctx corectx.Context, orderId string) (*GatewayOrderStatus, error)
}

// CreateGatewayPaymentCommand asks for a collection to be opened.
type CreateGatewayPaymentCommand struct {
	// OrgId owns the money. Passed explicitly because paymentinvoice writes it on every record and
	// will not guess which organization a multi-org caller means.
	OrgId string

	PaymentMethodId string
	Amount          decimal.Decimal

	// Content is what the payer sees on their statement.
	Content string

	// SalesPaymentId, SalesBillId travel as the order's metadata and come back on the settlement, so
	// a verdict can be matched to the payment awaiting it even if the correlation column were lost.
	SalesPaymentId string
	SalesBillId    string
}

// CreateGatewayPaymentResultData is what a till needs to put in front of the customer.
type CreateGatewayPaymentResultData struct {
	// OrderId is what Sales stores and quotes for a refund. OrderCode is the gateway's own key,
	// kept for reconciliation and support. Storing one where the other belongs means the order
	// cannot be refunded, so both are carried.
	OrderId   string
	OrderCode string

	// QrCodeUrl and PayUrl are both empty for a card terminal, where the prompt is pushed to the
	// device the customer is standing at.
	QrCodeUrl string
	PayUrl    string
}

// CreateGatewayPaymentResult carries a refusal the till can act on — a method withdrawn, an amount
// out of bounds, a gateway declining — in Refused/RefusalReason rather than as a Go error, matching
// how the upstream port reports the same things. An error means the request could not be processed.
type CreateGatewayPaymentResult struct {
	Data    CreateGatewayPaymentResultData
	HasData bool

	Refused       bool
	RefusalReason string
}

// RefundGatewayPaymentCommand gives money back against an order Sales opened.
type RefundGatewayPaymentCommand struct {
	OrderId string
	Amount  decimal.Decimal
	Content string
}

type RefundGatewayPaymentResultData struct {
	OrderId      string
	RefundAmount decimal.Decimal

	// RemainingAmount is what is left of the order after this refund, which Sales cross-checks
	// against its own remaining-refundable arithmetic. Two modules disagreeing about how much is
	// left is worth knowing about before the money moves again.
	RemainingAmount decimal.Decimal
}

type RefundGatewayPaymentResult struct {
	Data    RefundGatewayPaymentResultData
	HasData bool

	Refused       bool
	RefusalReason string
}

// GatewayOrderStatus is where an order stands, for reconciliation.
type GatewayOrderStatus struct {
	// Found is false when paymentinvoice has no such order. Distinguished from an error because it
	// is an answer: a payment quoting an order that does not exist is a broken record, not a
	// failure to read.
	Found bool

	OrderId string

	// Settled, Failed and Pending are the three answers reconciliation acts on. Sales does not
	// reproduce paymentinvoice's order states, because doing so would put another module's state
	// machine in this one where it would go stale; the adapter maps them here instead.
	Settled bool
	Failed  bool

	RefTransactionId string
	Amount           decimal.Decimal
	RefundAmount     decimal.Decimal
}

// Package event declares what Sales listens for on the internal event bus.
//
// The payload is redeclared here rather than imported from the publishing module, for the same
// reason interfaces/external declares narrowed local interfaces: the coupling between two modules
// should be the JSON contract, not a Go type. Importing paymentinvoice's struct would make every
// field it ever adds part of Sales' compile, and would tie the two modules together in a build
// where one of them may eventually be a separate process.
package event

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// PaymentSettledType is what became of a payment order. The values match the publisher's, which is
// the contract; the constants are named on both sides so neither has to quote a bare string.
type PaymentSettledType string

const (
	PaymentSettledPaid     PaymentSettledType = "paid"
	PaymentSettledFailed   PaymentSettledType = "failed"
	PaymentSettledExpired  PaymentSettledType = "expired"
	PaymentSettledCanceled PaymentSettledType = "canceled"
)

// PaymentSettledEvent is one payment order's verdict, as it arrives on the bus.
type PaymentSettledEvent struct {
	Type PaymentSettledType `json:"type"`

	// TenantId travels on the event because the subscriber runs in its own goroutine with no request
	// context. Empty in a single-tenant build.
	TenantId string `json:"tenant_id,omitempty"`

	OrgId string `json:"org_id,omitempty"`

	OrderId string `json:"order_id"`
	OrderPk string `json:"order_pk,omitempty"`

	// RefTransactionId is the gateway's identifier for the payment, present once money moved.
	RefTransactionId string `json:"ref_transaction_id,omitempty"`

	// Amount is a decimal as a string, never a JSON number: money that has passed through a float
	// is no longer the money that was taken.
	Amount string `json:"amount,omitempty"`

	// Metadata is what Sales attached when it opened the order, echoed back. It carries the
	// sales_payment_id that the correlation column would otherwise supply.
	Metadata map[string]any `json:"metadata,omitempty"`
}

// SalesPaymentIdFromMetadata reads the fallback correlation out of the echoed metadata.
//
// The payment order id is the normal way a verdict finds its payment; this covers the narrow case
// where the order was opened and the process died before that column was written.
func (this PaymentSettledEvent) SalesPaymentIdFromMetadata() string {
	if this.Metadata == nil {
		return ""
	}
	if value, ok := this.Metadata["sales_payment_id"].(string); ok {
		return value
	}
	return ""
}

// PaymentSettledHandler acts on one verdict.
type PaymentSettledHandler interface {
	Handle(ctx corectx.Context, event PaymentSettledEvent) error
}

// PaymentSettledHandlerRegistry dispatches by verdict, so that adding an outcome is a new entry
// rather than another branch inside one handler.
type PaymentSettledHandlerRegistry map[PaymentSettledType]PaymentSettledHandler

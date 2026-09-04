// Package event is what this module announces to the rest of the build when an order reaches a
// verdict.
//
// It exists alongside the HTTP result sync, not instead of it. The sync tells one ordering system
// that asked to be called back; this tells whoever is listening in the same process, which is what a
// module in the same binary needs rather than a loopback HTTP request to itself.
//
// NEITHER IS A GUARANTEE. The bus acknowledges a message before a subscriber has handled it, so a
// crash in the wrong moment loses the announcement. A subscriber that must not miss a settlement
// reconciles as well, through OrderDomainService.GetOrderStatus — this is the fast path, and the
// order row is the truth.
package event

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// PaymentSettledType is what happened to the order.
//
// One topic carries all four and the type discriminates, rather than a topic per outcome: a
// subscriber cares about the same order whichever way it went, and splitting the topics would make
// it subscribe four times to learn about one payment.
type PaymentSettledType string

const (
	// PaymentSettledPaid — the money is in.
	PaymentSettledPaid PaymentSettledType = "paid"

	// PaymentSettledFailed — the gateway refused, or the payer's attempt failed.
	PaymentSettledFailed PaymentSettledType = "failed"

	// PaymentSettledExpired — nobody paid within the window and the watchdog closed the order.
	PaymentSettledExpired PaymentSettledType = "expired"

	// PaymentSettledCanceled — the order was called off before it was paid.
	PaymentSettledCanceled PaymentSettledType = "canceled"
)

// PaymentSettledEvent announces one order's verdict.
type PaymentSettledEvent struct {
	Type PaymentSettledType `json:"type"`

	// TenantId travels ON THE EVENT because the subscriber runs in its own goroutine with no request
	// context: it is the only way it learns whose data to read. Empty in a single-tenant build.
	TenantId string `json:"tenant_id,omitempty"`

	OrgId string `json:"org_id,omitempty"`

	// OrderId is the identifier the ordering system was given and stores. OrderPk is this module's
	// own primary key, carried so a subscriber that kept it need not look the order up again.
	OrderId string `json:"order_id"`
	OrderPk string `json:"order_pk,omitempty"`

	// RefTransactionId is the gateway's identifier for the payment, present once money moved. Empty
	// on a failure or an expiry, where no transaction was ever completed.
	RefTransactionId string `json:"ref_transaction_id,omitempty"`

	// Amount is a decimal as a STRING. Serialising it as a JSON number would route it through a
	// float on the way back in, and money that has been through a float is no longer the money that
	// was taken.
	Amount string `json:"amount,omitempty"`

	// Metadata is what the caller attached when the order was opened, echoed back untouched. It is
	// how a subscriber matches a verdict to whatever it is holding: the correlation belongs to the
	// caller, and this module neither reads nor interprets it.
	Metadata map[string]any `json:"metadata,omitempty"`
}

// PaymentSettledEventPublisher announces order verdicts.
//
// Only PublishAsync is offered, deliberately. Every caller here is on the far side of a committed
// settlement — the order reached its verdict in its own transaction before this is reached — so a
// slow or absent broker must never fail the work that already succeeded. A synchronous variant
// would exist only to be misused at exactly the point where it does the most harm.
type PaymentSettledEventPublisher interface {
	PublishAsync(ctx corectx.Context, event PaymentSettledEvent)
}

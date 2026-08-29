// Package gateway declares what a payment gateway must be able to do, and nothing about how any
// particular one does it.
//
// The gateways this module talks to agree on very little. MoMo returns a QR code and a pay URL and
// settles through an IPN callback; a card terminal returns neither and settles when the cashier
// finishes with the customer standing there; VietQR has the bank call us and expects a reply shaped
// unlike anyone else's. They disagree on what an order must carry, too: a terminal payment is
// meaningless without knowing which terminal, and a wallet payment has no such notion.
//
// So the port is deliberately narrow, and every difference is pushed behind it. An adapter states
// what it needs, checks it itself, and writes what it wants to keep — no code outside the adapter
// package branches on which gateway is in play. Adding a gateway is a new package implementing
// this interface plus a payment-method row naming it; it is not an edit to the order service.
package gateway

import (
	"github.com/shopspring/decimal"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// PaymentGateway is one gateway's implementation of the payment lifecycle.
//
// Implementations must be safe for concurrent use: one instance serves every request for its
// adapter code.
type PaymentGateway interface {
	// AdapterCode identifies this implementation. A payment-method row selects an adapter by
	// storing this value, so it is a wire identifier and may not change once deployed.
	AdapterCode() string

	// ValidateOrder checks the request against what this gateway requires, before anything is
	// written or any money is asked for.
	//
	// It is where a gateway states its own preconditions: the card-terminal adapter rejects an
	// order that names no terminal, a wallet adapter has nothing to add. Violations are appended
	// to vErrs and become a 400 naming the offending field; a returned error means the check
	// itself could not be carried out, which is a 500.
	ValidateOrder(ctx corectx.Context, req OrderRequest, vErrs *ft.ClientErrors) error

	// PrepareMetadata returns what should be stored on the order for this gateway, given the
	// caller's input. It runs after ValidateOrder has passed, so it may assume its requirements
	// are met.
	//
	// Returning the map rather than mutating the order keeps the adapter from deciding when a
	// write happens: the domain service merges it inside the same transaction as the order.
	PrepareMetadata(ctx corectx.Context, req OrderRequest) (map[string]any, error)

	// CreatePayment asks the gateway to start collecting.
	CreatePayment(ctx corectx.Context, req CreatePaymentRequest) (*CreatePaymentResult, error)

	// Refund returns money for a payment this gateway collected.
	Refund(ctx corectx.Context, req RefundRequest) (*RefundResult, error)

	// CheckOrder asks the gateway what became of a payment, for orders that never received a
	// callback. It is what stops an order sitting in limbo when a callback is lost.
	CheckOrder(ctx corectx.Context, req CheckOrderRequest) (*CheckOrderResult, error)
}

// OrderRequest is what the caller asked for, before an order exists.
type OrderRequest struct {
	Amount       decimal.Decimal
	CurrencyCode string
	Content      *string

	// Metadata is the method-specific input the caller supplied, uninterpreted. Each adapter
	// reads the keys it declares and ignores the rest.
	Metadata map[string]any

	// MethodConfig is the non-secret configuration of the payment-method row in play, which is
	// how one adapter serves two merchant accounts.
	MethodConfig map[string]any

	// ProfileConfig is the decrypted credentials of the payment profile this order is collected
	// through, or nil when it names none.
	//
	// It is separate from MethodConfig because the two answer different questions and carry
	// different risk. A method config says how a gateway behaves and is readable through the API;
	// a profile config is the merchant secret and is encrypted at rest. An adapter overlays the
	// credentials it recognizes onto the ones it was built with, so a deployment that configures
	// one merchant account keeps working untouched and a profile only overrides what it supplies.
	ProfileConfig map[string]any
}

// CreatePaymentRequest is a persisted order handed to its gateway.
type CreatePaymentRequest struct {
	OrderRequest

	// OrderCode is the identifier the gateway knows this order by, and the key its callback
	// will arrive under. Never the order's id or its quoted order_id.
	OrderCode string
}

// CreatePaymentResult is what the payer needs in order to pay.
//
// Both URLs are empty for a card terminal: there the prompt is pushed to the device and the
// customer is standing at it, so there is nothing to show on a screen here.
type CreatePaymentResult struct {
	QrCodeUrl string
	PayUrl    string

	// RefTransactionId is the gateway's own identifier, when it issues one at create time.
	RefTransactionId string

	// RawResponse is the reply as received, kept as the evidence for what was asked and answered.
	RawResponse map[string]any
}

type RefundRequest struct {
	OrderCode string
	Amount    decimal.Decimal

	// Content is why the money is being returned. Some gateways require a description on a
	// refund even though they do not on a payment.
	Content *string

	// RefTransactionId identifies the payment being reversed, as the gateway knows it.
	RefTransactionId string

	Metadata     map[string]any
	MethodConfig map[string]any

	// ProfileConfig is the decrypted credentials of the payment profile the order was collected
	// through, or nil when it names none. See OrderRequest.ProfileConfig.
	ProfileConfig map[string]any
}

type RefundResult struct {
	RefTransactionId string
	RawResponse      map[string]any
}

type CheckOrderRequest struct {
	OrderCode string

	// Amount is the order's own amount. Some gateways will not answer a status query without it.
	Amount decimal.Decimal

	Metadata     map[string]any
	MethodConfig map[string]any

	// ProfileConfig is the decrypted credentials of the payment profile the order was collected
	// through, or nil when it names none. See OrderRequest.ProfileConfig.
	ProfileConfig map[string]any
}

// CheckOrderResult is the gateway's verdict on an order it was asked about.
//
// Settled distinguishes "the gateway has reached an outcome" from what that outcome was: an
// unsettled order is still in flight and must be left alone, while a settled one is Paid or not.
type CheckOrderResult struct {
	Settled          bool
	Paid             bool
	RefTransactionId string
	RawResponse      map[string]any
}

// ProfileString reads one credential out of a payment profile's decrypted config.
//
// The config is written by an operator rather than by this codebase, and the same credential is
// spelled differently depending on which console wrote it — the service this module supersedes
// used camelCase throughout, while everything on this side is snake_case. Accepting both here is
// what stops a profile that reads correctly in one system from silently falling back to the
// deployment's own credentials in the other. The first key present wins.
//
// A missing key answers "", which callers read as "the profile does not override this".
func ProfileString(config map[string]any, keys ...string) string {
	if config == nil {
		return ""
	}

	for _, key := range keys {
		value, exists := config[key]
		if !exists || value == nil {
			continue
		}
		if text, isText := value.(string); isText && text != "" {
			return text
		}
	}
	return ""
}

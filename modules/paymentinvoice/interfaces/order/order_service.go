package order

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// OrderDomainService is what another module calls to move money through this one.
//
// Both operations report a rule the caller broke through the result's ClientErrors, never as a Go
// error — a gateway refusing a payment, an order that cannot be refunded twice, a payment method
// withdrawn from use are all things the caller can act on. A returned error means the request
// could not be processed at all.
//
// There is deliberately no application-service counterpart. Authorization for a cross-module call
// is already established by the request that started it, so callers reach the domain service
// directly, which is the convention every other module's interfaces/external alias follows.
type OrderDomainService interface {
	CreatePayment(ctx corectx.Context, cmd CreatePaymentCommand) (*CreatePaymentResult, error)
	Refund(ctx corectx.Context, cmd RefundCommand) (*RefundResult, error)

	// GetOrderStatus reads back where an order stands, for a caller reconciling its own record of a
	// payment against this module's.
	//
	// It exists because settlement is announced, not returned: an order goes out to a gateway and
	// comes back through a callback or the watchdog, and any announcement can be missed. A caller
	// holding a payment that has been pending too long asks here rather than waiting forever.
	//
	// Read-only, and deliberately not a way to fetch the order row: a caller has no business with
	// the merchant profile or the sync logs, so this answers only what reconciliation needs.
	GetOrderStatus(ctx corectx.Context, query GetOrderStatusQuery) (*GetOrderStatusResult, error)
}

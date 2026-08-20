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
}

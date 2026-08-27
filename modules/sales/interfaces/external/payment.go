// Package external declares Sales' local ports onto the capabilities other modules offer.
//
// Sales code depends on these interfaces and never on another module's service directly, so that
// splitting a module into its own process changes only the binding in infra/external. See
// docs/wiki/01 "Microservice-ready Monolith".
//
// They are narrowed local interfaces rather than aliases of the upstream service. The difference
// matters: an alias would re-export every method the upstream adds later, so Sales would silently
// gain the ability to reach into another module's data the day that module grew a method.
package external

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itMethod "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/paymentmethod"
)

// PaymentMethodExtService is Sales' port onto paymentinvoice's payment-method master data.
//
// It exposes reading and judging, and no writing at all: a payment method is paymentinvoice's
// record, and Sales holds only a mapping saying which of them a channel accepts. The port
// deliberately offers no way to create or edit one, so that a bug in Sales cannot write into
// another module's master data.
//
// Note what is NOT here: any way to ask for the raw row. Usability is not derivable from the fields
// — three independent gates apply, one of which depends on the running build — so the upstream
// answers the question rather than handing over the inputs to it. See the package comment on
// paymentinvoice/interfaces/paymentmethod (D-05a).
type PaymentMethodExtService interface {
	// ListPaymentMethods answers the methods this deployment knows about.
	//
	// Sales calls it with UsableOnly false when merging against its own mappings, because a mapped
	// method that has since become unusable must still appear — marked stale — rather than vanish
	// from a screen an administrator is using to fix exactly that (CR §34).
	ListPaymentMethods(
		ctx corectx.Context, query ListPaymentMethodsQuery,
	) (*ListPaymentMethodsResult, error)

	// AssertUsable answers whether one method may take a payment.
	//
	// This is the validation CR §31 requires before a mapping is written: enabling a method that
	// paymentinvoice cannot serve would create configuration that fails only at the moment a
	// customer tries to pay.
	AssertUsable(ctx corectx.Context, query AssertUsableQuery) (*AssertUsableResult, error)
}

type ListPaymentMethodsQuery = itMethod.ListPaymentMethodsQuery
type ListPaymentMethodsResult = itMethod.ListPaymentMethodsResult
type AssertUsableQuery = itMethod.AssertUsableQuery
type AssertUsableResult = itMethod.AssertUsableResult
type PaymentMethodData = itMethod.PaymentMethodData

// The unusability reasons, re-exported so that Sales names them without importing paymentinvoice
// outside infra/external.
const (
	ReasonInactive           = itMethod.ReasonInactive
	ReasonGatewayUnavailable = itMethod.ReasonGatewayUnavailable
	ReasonArchived           = itMethod.ReasonArchived
	ReasonAmountOutOfBounds  = itMethod.ReasonAmountOutOfBounds
)

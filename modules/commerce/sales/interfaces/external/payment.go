// Package external declares Sales' local ports onto the capabilities other modules offer, so
// splitting a module into its own process changes only the binding in infra/external. They are
// narrowed local interfaces rather than aliases: an alias would re-export every method the upstream
// adds later, silently widening what Sales can reach.
package external

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itMethod "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/paymentmethod"
)

// PaymentMethodExtService is Sales' port onto paymentinvoice's payment-method master data. It reads
// and judges but never writes, so a bug in Sales cannot alter another module's master data. There is
// also no way to fetch the raw row: usability is not derivable from the fields — three independent
// gates apply, one depending on the running build — so the upstream answers the question instead.
type PaymentMethodExtService interface {
	// ListPaymentMethods answers the methods this deployment knows about. Sales calls it with
	// UsableOnly false when merging against its own mappings, so a mapped method that became
	// unusable still appears — marked stale — on the screen an administrator uses to fix it.
	ListPaymentMethods(
		ctx corectx.Context, query ListPaymentMethodsQuery,
	) (*ListPaymentMethodsResult, error)

	// AssertUsable answers whether one method may take a payment. It runs before a mapping is
	// written, since enabling a method paymentinvoice cannot serve would fail only when a customer
	// tries to pay.
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

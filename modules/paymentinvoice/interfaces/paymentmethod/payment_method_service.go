// Package paymentmethod is the port through which another module reads this one's payment method
// master data.
//
// It exists because "is this method usable" is not derivable from the row. Three independent gates
// apply and only one of them is stored: the row's own is_active flag (a nil counts as inactive),
// the presence of its adapter_code in the gateway registry of *the running build* — which makes
// usability deployment-dependent — and the amount bounds, whose upper limit is exclusive while the
// lower is inclusive. A consumer re-deriving that from the row would be wrong on the day any of the
// three changed, and would be silently wrong.
//
// So the answer is computed here, beside the rules, and handed over as a decision rather than as
// the inputs to one.
package paymentmethod

import (
	"github.com/shopspring/decimal"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// PaymentMethodAppService is the published capability. Consumers depend on this interface, never on
// paymentinvoice's domain models or its resource engine.
type PaymentMethodAppService interface {
	// ListPaymentMethods answers the methods this deployment knows about, each carrying whether it
	// is presently usable and why not when it is not.
	ListPaymentMethods(
		ctx corectx.Context, query ListPaymentMethodsQuery,
	) (*ListPaymentMethodsResult, error)

	// AssertUsable answers whether one method may take a payment, optionally of a given amount.
	//
	// A refusal is a client error rather than a Go error: the caller named a method that exists but
	// cannot be used, which is a thing they can act on.
	AssertUsable(ctx corectx.Context, query AssertUsableQuery) (*AssertUsableResult, error)
}

// ListPaymentMethodsQuery narrows the listing.
type ListPaymentMethodsQuery struct {
	// UsableOnly drops the methods that cannot presently take a payment. Callers building a
	// customer-facing chooser want this; callers reconciling their own stored configuration against
	// the master list do not, because a method they have mapped and which has since become unusable
	// must still appear, marked, rather than vanish.
	UsableOnly bool
}

// AssertUsableQuery names the method and, when the caller has one, the amount.
type AssertUsableQuery struct {
	PaymentMethodId string

	// Amount is optional. When nil the bounds are not checked and the answer covers only the row
	// and the adapter — which is what a configuration screen needs, since it is deciding whether a
	// method may ever be offered rather than whether one particular payment may proceed.
	Amount *decimal.Decimal
}

// PaymentMethodData is the flattened view a consumer receives. It deliberately carries no
// adapter_code: which adapter serves a method is this module's business, and a consumer that
// branched on it would be encoding a payment integration it does not own.
type PaymentMethodData struct {
	Id   string `json:"id"`
	Code string `json:"code"`
	// Name is the multilingual value as stored, not a string flattened to one language. There is
	// no house rule for choosing a language here, and picking one would bake this module's guess
	// into every consumer's UI; the caller knows the requester's locale and this module does not.
	Name      model.LangJson   `json:"name,omitempty"`
	IsActive  bool             `json:"is_active"`
	IsUsable  bool             `json:"is_usable"`
	MinAmount *decimal.Decimal `json:"min_amount,omitempty"`

	// MaxAmount is EXCLUSIVE: an amount equal to it is refused. The name does not say so, so the
	// comment must — a consumer treating it as inclusive would offer a method that then rejects the
	// payment at the last step.
	MaxAmount *decimal.Decimal `json:"max_amount,omitempty"`

	// UnusableReason is empty when IsUsable, and otherwise names which gate closed, so that a
	// consumer can tell an administrator what to fix rather than only that something is wrong.
	UnusableReason string `json:"unusable_reason,omitempty"`

	// HasGateway says whether taking this payment means opening an order with a provider and waiting
	// for it to settle, or whether the money is simply in — cash across a counter.
	//
	// It is a capability, NOT the adapter's name, and the distinction is the whole point: a consumer
	// has to know which of its two flows to run, but naming momo or vietqr here would let it branch
	// on a payment integration it does not own, which is what this shape otherwise refuses to allow.
	HasGateway bool `json:"has_gateway"`
}

type ListPaymentMethodsResult struct {
	ClientErrors ft.ClientErrors     `json:"client_errors,omitempty"`
	Data         []PaymentMethodData `json:"data"`
	HasData      bool                `json:"has_data"`
}

type AssertUsableResult struct {
	ClientErrors ft.ClientErrors   `json:"client_errors,omitempty"`
	Data         PaymentMethodData `json:"data"`
	HasData      bool              `json:"has_data"`
}

// The reasons AssertUsable and ListPaymentMethods report. They are the four ways a method can fail
// to be usable, and they are named rather than described so that a consumer can branch on them.
const (
	// ReasonInactive — the row's is_active is false or absent.
	ReasonInactive = "inactive"

	// ReasonGatewayUnavailable — the row names an adapter this build does not ship or has not
	// enabled. Configuration is correct; the deployment is missing something.
	ReasonGatewayUnavailable = "gateway_unavailable"

	// ReasonArchived — the row is archived. Distinct from inactive: archiving retires a method
	// permanently, deactivating pauses it.
	ReasonArchived = "archived"

	// ReasonAmountOutOfBounds — the method is fine but this amount is not one it accepts.
	ReasonAmountOutOfBounds = "amount_out_of_bounds"
)

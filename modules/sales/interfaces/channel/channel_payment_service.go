package channel

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// ChannelPaymentAppService is the payment-method configuration of a sales channel.
//
// It is a capability of the channel rather than a resource of its own, which is why it lives here
// beside SalesChannelAppService and not behind a REST resource: a mapping row has no identity a
// client would name, and the only sensible operations on it are "this channel accepts this method"
// and "it no longer does".
type ChannelPaymentAppService interface {
	// ListChannelPaymentMethods answers the merged view: every payment method paymentinvoice knows
	// about, each marked with whether this channel accepts it (CR §29).
	//
	// The merge happens here and not in the browser. That is the requirement, and the reason is
	// that the join needs facts neither side holds alone — a method enabled here but absent
	// upstream is stale, and only a reader holding both lists can see that. A frontend joining two
	// module APIs would have to reimplement that rule, and would get it wrong the first time
	// either endpoint failed.
	ListChannelPaymentMethods(
		ctx corectx.Context, query ListChannelPaymentMethodsQuery,
	) (*ListChannelPaymentMethodsResult, error)

	// EnableChannelPaymentMethod records that a channel accepts a payment method (CR §31).
	//
	// The method is validated through the paymentinvoice port first: enabling one this deployment
	// cannot serve would create configuration that fails only when a customer tries to pay.
	// Idempotent — enabling an already-enabled method succeeds.
	EnableChannelPaymentMethod(
		ctx corectx.Context, command ChannelPaymentMethodCommand,
	) (*ChannelPaymentMutationResult, error)

	// DisableChannelPaymentMethod removes the mapping (CR §32). Idempotent, and deliberately does
	// NOT validate the method: a mapping whose method has vanished upstream must stay removable,
	// since removing it is the fix an administrator is applying.
	DisableChannelPaymentMethod(
		ctx corectx.Context, command ChannelPaymentMethodCommand,
	) (*ChannelPaymentMutationResult, error)

	// IsPaymentMethodEnabledForChannel is the enforcement query SALES-027 calls at payment time.
	//
	// It reports the mapping only, never usability: those are two different questions with two
	// different owners, and a caller taking a payment must ask both. Default-deny — a channel with
	// no mappings accepts nothing (CR §76).
	IsPaymentMethodEnabledForChannel(
		ctx corectx.Context, query IsPaymentMethodEnabledQuery,
	) (*IsPaymentMethodEnabledResult, error)
}

// ListChannelPaymentMethodsQuery names the channel. Either identifier will do — a REST caller has
// the id, an integrating module has only the code.
type ListChannelPaymentMethodsQuery struct {
	SalesChannelId   string
	SalesChannelCode string

	// EnabledOnly narrows the answer to what this channel actually accepts, which is what a
	// checkout screen wants. An administration screen leaves it false, because it is choosing from
	// the full list.
	EnabledOnly bool
}

// ChannelPaymentMethodCommand names one mapping.
type ChannelPaymentMethodCommand struct {
	SalesChannelId   string
	SalesChannelCode string
	PaymentMethodId  string
}

type IsPaymentMethodEnabledQuery struct {
	SalesChannelId   string
	SalesChannelCode string
	PaymentMethodId  string
}

// ChannelPaymentMethodData is one row of the merged view.
type ChannelPaymentMethodData struct {
	PaymentMethodId string         `json:"payment_method_id"`
	Code            string         `json:"code"`
	Name            model.LangJson `json:"name,omitempty"`

	// IsEnabled is Sales' half: does a mapping row exist for this channel.
	IsEnabled bool `json:"is_enabled"`

	// IsUsable is paymentinvoice's half: could this method take a payment at all. The two are
	// independent — a method can be enabled here and unusable there, which is exactly the state an
	// administrator needs to see rather than have hidden.
	IsUsable bool `json:"is_usable"`

	// UnusableReason names which gate closed upstream, empty when usable.
	UnusableReason string `json:"unusable_reason,omitempty"`

	// IsStale marks a mapping this channel holds for a method paymentinvoice no longer reports at
	// all (CR §34). It is not the same as unusable: unusable means the method exists and is
	// refused, stale means it has gone. A stale mapping blocks new payments, stays disable-able,
	// and is never silently deleted — the historical payments made through it must keep making
	// sense.
	IsStale bool `json:"is_stale"`
}

type ListChannelPaymentMethodsResult struct {
	ClientErrors ft.ClientErrors            `json:"client_errors,omitempty"`
	Data         []ChannelPaymentMethodData `json:"data"`
	HasData      bool                       `json:"has_data"`
}

type ChannelPaymentMutationResult struct {
	ClientErrors ft.ClientErrors `json:"client_errors,omitempty"`
	HasData      bool            `json:"has_data"`
}

type IsPaymentMethodEnabledResult struct {
	ClientErrors ft.ClientErrors `json:"client_errors,omitempty"`
	Data         bool            `json:"data"`
	HasData      bool            `json:"has_data"`
}

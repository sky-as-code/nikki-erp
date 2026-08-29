package channel

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// ChannelPaymentAppService is the payment-method configuration of a sales channel. It is a
// capability of the channel rather than a REST resource of its own: a mapping row has no identity a
// client would name.
type ChannelPaymentAppService interface {
	// ListChannelPaymentMethods answers the merged view: every payment method paymentinvoice knows
	// about, each marked with whether this channel accepts it. The merge happens here rather than
	// in the browser because staleness — enabled here but absent upstream — is only visible to a
	// reader holding both lists.
	ListChannelPaymentMethods(
		ctx corectx.Context, query ListChannelPaymentMethodsQuery,
	) (*ListChannelPaymentMethodsResult, error)

	// EnableChannelPaymentMethod records that a channel accepts a payment method. The method is
	// validated through the paymentinvoice port first, since enabling one this deployment cannot
	// serve would fail only when a customer tries to pay. Idempotent.
	EnableChannelPaymentMethod(
		ctx corectx.Context, command ChannelPaymentMethodCommand,
	) (*ChannelPaymentMutationResult, error)

	// DisableChannelPaymentMethod removes the mapping. Idempotent, and deliberately does not
	// validate the method: a mapping whose method vanished upstream must stay removable.
	DisableChannelPaymentMethod(
		ctx corectx.Context, command ChannelPaymentMethodCommand,
	) (*ChannelPaymentMutationResult, error)

	// IsPaymentMethodEnabledForChannel is the enforcement query at payment time. It reports the
	// mapping only, never usability — two questions with two owners, and a caller taking a payment
	// must ask both. Default-deny: a channel with no mappings accepts nothing.
	IsPaymentMethodEnabledForChannel(
		ctx corectx.Context, query IsPaymentMethodEnabledQuery,
	) (*IsPaymentMethodEnabledResult, error)
}

// ListChannelPaymentMethodsQuery names the channel. Either identifier will do — a REST caller has
// the id, an integrating module has only the code.
type ListChannelPaymentMethodsQuery struct {
	SalesChannelId   string
	SalesChannelCode string

	// EnabledOnly narrows the answer to what this channel accepts, which is what a checkout screen
	// wants; an administration screen leaves it false to choose from the full list.
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

	// IsUsable is paymentinvoice's half: could this method take a payment at all. Independent of
	// IsEnabled — a method can be enabled here and unusable there.
	IsUsable bool `json:"is_usable"`

	// UnusableReason names which gate closed upstream, empty when usable.
	UnusableReason string `json:"unusable_reason,omitempty"`

	// IsStale marks a mapping for a method paymentinvoice no longer reports at all — unusable means
	// the method exists and is refused, stale means it has gone. A stale mapping blocks new
	// payments, stays disable-able, and is never silently deleted, so historical payments made
	// through it keep making sense.
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

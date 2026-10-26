package channel

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type PointPaymentAppService interface {
	ListPointPaymentMethods(
		ctx corectx.Context, query ListPointPaymentMethodsQuery,
	) (*ListPointPaymentMethodsResult, error)

	EnablePointPaymentMethod(
		ctx corectx.Context, command PointPaymentMethodCommand,
	) (*PointPaymentMutationResult, error)

	DisablePointPaymentMethod(
		ctx corectx.Context, command PointPaymentMethodCommand,
	) (*PointPaymentMutationResult, error)
}

type PointPaymentMethodExtService interface {
	PayableMethodsOfPoint(
		ctx corectx.Context, query PayableMethodsQuery,
	) (*PayableMethodsResult, error)
}

type ListPointPaymentMethodsQuery struct {
	SalesPointId string

	EnabledOnly bool
}

type PointPaymentMethodCommand struct {
	SalesPointId     string
	PaymentMethodId  string
	PaymentProfileId string
}

type PayableMethodsQuery struct {
	SalesPointId string
}

type PointPaymentMethodData struct {
	PaymentMethodId  string         `json:"payment_method_id"`
	Code             string         `json:"code"`
	Name             model.LangJson `json:"name,omitempty"`
	PaymentProfileId string         `json:"payment_profile_id,omitempty"`

	IsEnabled bool `json:"is_enabled"`

	IsEnabledForChannel bool `json:"is_enabled_for_channel"`

	IsUsable bool `json:"is_usable"`

	UnusableReason string `json:"unusable_reason,omitempty"`

	IsStale bool `json:"is_stale"`
}

type ListPointPaymentMethodsResult struct {
	ClientErrors ft.ClientErrors          `json:"client_errors,omitempty"`
	Data         []PointPaymentMethodData `json:"data"`
	HasData      bool                     `json:"has_data"`
}

type PointPaymentMutationResult struct {
	ClientErrors ft.ClientErrors `json:"client_errors,omitempty"`
	HasData      bool            `json:"has_data"`
}

type PayableMethod struct {
	Id   string         `json:"id"`
	Code string         `json:"code"`
	Name model.LangJson `json:"name,omitempty"`
}

type PayableMethodsResult struct {
	ClientErrors ft.ClientErrors `json:"client_errors,omitempty"`
	Methods      []PayableMethod `json:"methods"`
	HasData      bool            `json:"has_data"`
}

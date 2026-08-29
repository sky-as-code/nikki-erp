// Package channel is Sales' outward-facing contract for sales channels and sales points. Another
// module depends on this interface and never on the application service directly, so splitting Sales
// into its own process changes only the binding, not the callers.
package channel

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// SalesChannelAppService is the capability set another module needs from Sales. Listing, editing
// and archiving channels are absent deliberately: they belong to a human at the REST surface, and a
// module that could archive a channel could stop a business selling as a deployment side effect.
type SalesChannelAppService interface {
	// RegisterSalesChannel claims a channel for the calling module, idempotently by code. A code
	// already owned by a different module is refused rather than taken over.
	RegisterSalesChannel(
		ctx corectx.Context, command RegisterSalesChannelCommand,
	) (*RegisterSalesChannelResult, error)

	// ResolveSalesChannelByCode turns a stable integration code into the id and current state of
	// the channel it names, so no caller has to store a database id.
	ResolveSalesChannelByCode(
		ctx corectx.Context, query ResolveSalesChannelQuery,
	) (*ResolveSalesChannelResult, error)
}

// SalesPointAppService is the sales point half of the same contract.
type SalesPointAppService interface {
	// CreateSalesPoint registers one selling place under a channel, idempotently by the pair
	// (channel, external reference id). A caller retrying after a timeout is handed the point it
	// created the first time, so creating a kiosk and its sales point need no distributed
	// transaction.
	CreateSalesPoint(
		ctx corectx.Context, command CreateSalesPointCommand,
	) (*CreateSalesPointResult, error)

	// ArchiveSalesPoint retires a sales point. Idempotent.
	ArchiveSalesPoint(
		ctx corectx.Context, command SalesPointCommand,
	) (*SalesPointMutationResult, error)

	// SuspendSalesPoint stops a sales point taking new orders, leaving returns and refunds working.
	SuspendSalesPoint(
		ctx corectx.Context, command SalesPointCommand,
	) (*SalesPointMutationResult, error)

	// ActivateSalesPoint returns a suspended sales point to service.
	ActivateSalesPoint(
		ctx corectx.Context, command SalesPointCommand,
	) (*SalesPointMutationResult, error)

	// DeleteSalesPoint removes a sales point, or archives it when it carries sales history.
	// The result says which happened.
	DeleteSalesPoint(
		ctx corectx.Context, command SalesPointCommand,
	) (*DeleteSalesPointResult, error)
}

type RegisterSalesChannelCommand struct {
	Code            string `json:"code"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	ManagedByModule string `json:"managed_by_module"`
}

type SalesChannelData struct {
	Id   string `json:"id"`
	Code string `json:"code"`
}

type RegisterSalesChannelResult struct {
	ClientErrors ft.ClientErrors  `json:"client_errors,omitempty"`
	Data         SalesChannelData `json:"data"`
	HasData      bool             `json:"has_data"`
}

type ResolveSalesChannelQuery struct {
	Code string `json:"code"`
}

type ResolvedSalesChannel struct {
	Id              string `json:"id"`
	Code            string `json:"code"`
	Name            string `json:"name"`
	ManagedByModule string `json:"managed_by_module"`
	Status          string `json:"status"`
	// IsUsable folds status and archive state together: a caller checking only one would let a
	// suspended channel trade.
	IsUsable bool `json:"is_usable"`
}

type ResolveSalesChannelResult struct {
	ClientErrors ft.ClientErrors      `json:"client_errors,omitempty"`
	Data         ResolvedSalesChannel `json:"data"`
	HasData      bool                 `json:"has_data"`
}

type CreateSalesPointCommand struct {
	// SalesChannelCode names the channel by its stable code rather than its id, so the caller
	// stores no database identifier of ours.
	SalesChannelCode string `json:"sales_channel_code"`
	Name             string `json:"name"`
	Code             string `json:"code"`
	// ExternalReferenceId is the caller's own id for the thing this point represents, and the key
	// a retry resolves against.
	ExternalReferenceId string `json:"external_reference_id"`
	// ExternalReferenceType says what kind of record that id names, as "{module}.{resource}".
	ExternalReferenceType string `json:"external_reference_type"`
}

type SalesPointData struct {
	Id               string `json:"id"`
	SalesChannelId   string `json:"sales_channel_id"`
	SalesChannelCode string `json:"sales_channel_code"`
	// AlreadyExisted tells a caller its retry was recognised rather than a new point created.
	AlreadyExisted bool `json:"already_existed"`
}

type CreateSalesPointResult struct {
	ClientErrors ft.ClientErrors `json:"client_errors,omitempty"`
	Data         SalesPointData  `json:"data"`
	HasData      bool            `json:"has_data"`
}

type SalesPointCommand struct {
	SalesPointId string `json:"sales_point_id"`
}

type SalesPointMutationResult struct {
	ClientErrors ft.ClientErrors `json:"client_errors,omitempty"`
	HasData      bool            `json:"has_data"`
}

type DeleteSalesPointResult struct {
	ClientErrors ft.ClientErrors `json:"client_errors,omitempty"`
	// Archived is true when sales history forced the point to be retired instead of removed.
	Archived bool `json:"archived"`
	HasData  bool `json:"has_data"`
}

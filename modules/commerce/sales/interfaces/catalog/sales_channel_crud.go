package catalog

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// SalesChannelRepository reads and writes sales channels.
type SalesChannelRepository interface {
	dynamicresource.CrudRepository
}

// SalesChannelDomainService adds the business lifecycle to the default CRUD. Suspension is
// distinct from archiving: suspended stops new sales points and new orders while leaving history,
// returns, refunds and fiscal adjustments working, and the two must not substitute for each other.
type SalesChannelDomainService interface {
	dynamicresource.CrudDomainService

	// ResolveByCode answers the channel a caller named by its stable code rather than its id.
	ResolveByCode(ctx corectx.Context, code string, orgId string) (*dyn.OpResult[dyn.MutateResultData], error)

	Suspend(ctx corectx.Context, channelId string) (*dyn.OpResult[dyn.MutateResultData], error)
	Activate(ctx corectx.Context, channelId string) (*dyn.OpResult[dyn.MutateResultData], error)

	// Archive is the system lifecycle, and refuses while the channel still has active sales points.
	Archive(ctx corectx.Context, channelId string) (*dyn.OpResult[dyn.MutateResultData], error)

	// AssertMutable is the gate the operations that attach new business to a channel go through.
	AssertMutable(ctx corectx.Context, channelId string) (*dyn.OpResult[dyn.MutateResultData], error)
}

// SalesChannelApplicationService is the authorized surface the REST handler serves.
type SalesChannelApplicationService interface {
	dynamicresource.CrudApplicationService

	Suspend(ctx corectx.Context, cmd ChannelLifecycleCommand) (*ChannelLifecycleResult, error)
	Activate(ctx corectx.Context, cmd ChannelLifecycleCommand) (*ChannelLifecycleResult, error)
	Archive(ctx corectx.Context, cmd ChannelLifecycleCommand) (*ChannelLifecycleResult, error)

	// Resolve is collection-level: it answers by code, so there is no record to address yet.
	Resolve(ctx corectx.Context, query ResolveChannelQuery) (*dyn.OpResult[any], error)

	// The payment-method routes bind their field map and delegate to ChannelPaymentAppService,
	// which owns the merge with the upstream catalogue and performs its own permission check. They
	// are declared here so the channel's REST handler has one service to talk to.
	PaymentMethods(ctx corectx.Context, query ChannelPaymentMethodsQuery) (*dyn.OpResult[any], error)
	EnablePaymentMethod(ctx corectx.Context, cmd ChannelPaymentMethodCommand) (*ChannelPaymentMethodResult, error)
	DisablePaymentMethod(ctx corectx.Context, cmd ChannelPaymentMethodCommand) (*ChannelPaymentMethodResult, error)
}

type (
	ChannelLifecycleCommand     = dynamicresource.UpdateCommand
	ChannelLifecycleResult      = dynamicresource.MutateResult
	ResolveChannelQuery         = dynamicresource.SearchQuery
	ChannelPaymentMethodsQuery  = dynamicresource.GetByIdQuery
	ChannelPaymentMethodCommand = dynamicresource.UpdateCommand
	ChannelPaymentMethodResult  = dynamicresource.MutateResult
)

type (
	CreateSalesChannelCommand      = dynamicresource.CreateCommand
	UpdateSalesChannelCommand      = dynamicresource.UpdateCommand
	DeleteSalesChannelCommand      = dynamicresource.DeleteCommand
	SetSalesChannelArchivedCommand = dynamicresource.SetArchivedCommand
	GetSalesChannelByIdQuery       = dynamicresource.GetByIdQuery
	SearchSalesChannelsQuery       = dynamicresource.SearchQuery
	SalesChannelExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateSalesChannelResult      = dynamicresource.CreateResult
	UpdateSalesChannelResult      = dynamicresource.MutateResult
	DeleteSalesChannelResult      = dynamicresource.MutateResult
	SetSalesChannelArchivedResult = dynamicresource.MutateResult
	GetSalesChannelByIdResult     = dynamicresource.GetOneResult
	SearchSalesChannelsResult     = dynamicresource.SearchResult
	SalesChannelExistsResult      = dynamicresource.ExistsResult
)

package catalog

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// SalesChannelRepository reads and writes sales channels.
type SalesChannelRepository interface {
	composable.CrudRepository
}

// SalesChannelDomainService adds the business lifecycle to the default CRUD. Suspension is
// distinct from archiving: suspended stops new sales points and new orders while leaving history,
// returns, refunds and fiscal adjustments working, and the two must not substitute for each other.
type SalesChannelDomainService interface {
	composable.CrudDomainService

	// ResolveByCode answers the channel a caller named by its stable code rather than its id.
	ResolveByCode(ctx corectx.Context, code string) (*dyn.OpResult[dyn.MutateResultData], error)

	Suspend(ctx corectx.Context, channelId string) (*dyn.OpResult[dyn.MutateResultData], error)
	Activate(ctx corectx.Context, channelId string) (*dyn.OpResult[dyn.MutateResultData], error)

	// Archive is the system lifecycle, and refuses while the channel still has active sales points.
	Archive(ctx corectx.Context, channelId string) (*dyn.OpResult[dyn.MutateResultData], error)

	// AssertMutable is the gate the operations that attach new business to a channel go through.
	AssertMutable(ctx corectx.Context, channelId string) (*dyn.OpResult[dyn.MutateResultData], error)
}

// SalesChannelApplicationService is the authorized surface the REST handler serves.
type SalesChannelApplicationService interface {
	composable.CrudApplicationService

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
	ChannelLifecycleCommand     = composable.UpdateCommand
	ChannelLifecycleResult      = composable.MutateResult
	ResolveChannelQuery         = composable.SearchQuery
	ChannelPaymentMethodsQuery  = composable.GetByIdQuery
	ChannelPaymentMethodCommand = composable.UpdateCommand
	ChannelPaymentMethodResult  = composable.MutateResult
)

type (
	CreateSalesChannelCommand      = composable.CreateCommand
	UpdateSalesChannelCommand      = composable.UpdateCommand
	DeleteSalesChannelCommand      = composable.DeleteCommand
	SetSalesChannelArchivedCommand = composable.SetArchivedCommand
	GetSalesChannelByIdQuery       = composable.GetByIdQuery
	SearchSalesChannelsQuery       = composable.SearchQuery
	SalesChannelExistsQuery        = composable.ExistsQuery
)

type (
	CreateSalesChannelResult      = composable.CreateResult
	UpdateSalesChannelResult      = composable.MutateResult
	DeleteSalesChannelResult      = composable.MutateResult
	SetSalesChannelArchivedResult = composable.MutateResult
	GetSalesChannelByIdResult     = composable.GetOneResult
	SearchSalesChannelsResult     = composable.SearchResult
	SalesChannelExistsResult      = composable.ExistsResult
)

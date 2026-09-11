package catalog

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// SalesFulfillmentMethodRepository reads and writes the fulfillment policy catalogue.
type SalesFulfillmentMethodRepository interface {
	composable.CrudRepository
}

// SalesFulfillmentMethodDomainService adds the archive lifecycle to the default CRUD. There is no
// suspend here, unlike a channel or a point: a method is either offered to new orders or it is
// not, and the fulfillments that already snapshotted it keep running either way.
type SalesFulfillmentMethodDomainService interface {
	composable.CrudDomainService

	// Archive withdraws a method from new orders, refusing while a channel or point still names it
	// as a default -- the configuration left behind would refuse every order it received.
	Archive(ctx corectx.Context, methodId string) (*dyn.OpResult[dyn.MutateResultData], error)

	// Unarchive returns a method to the selectable set.
	Unarchive(ctx corectx.Context, methodId string) (*dyn.OpResult[dyn.MutateResultData], error)

	// AssertAssignable is the gate every path that attaches a method to new business goes through:
	// order creation, and setting a channel or point default. channelId may be empty when the
	// caller is not yet bound to a channel.
	AssertAssignable(ctx corectx.Context, methodId string, channelId string) (*dyn.OpResult[dyn.MutateResultData], error)

	// IsAllowedOnChannel reports whether a channel permits a method. Default-deny: a channel with
	// no mapping rows permits nothing.
	IsAllowedOnChannel(ctx corectx.Context, channelId string, methodId string) (bool, error)
}

// SalesFulfillmentMethodApplicationService is the authorized surface the REST handler serves.
type SalesFulfillmentMethodApplicationService interface {
	composable.CrudApplicationService

	Archive(ctx corectx.Context, cmd ArchiveFulfillmentMethodCommand) (*ArchiveFulfillmentMethodResult, error)
	Unarchive(ctx corectx.Context, cmd ArchiveFulfillmentMethodCommand) (*ArchiveFulfillmentMethodResult, error)
}

// Archiving and unarchiving share one command shape and one permission: they are the same power in
// reverse, so whoever may retire a policy may undo it.
type (
	ArchiveFulfillmentMethodCommand = composable.UpdateCommand
	ArchiveFulfillmentMethodResult  = composable.MutateResult
)

type (
	CreateSalesFulfillmentMethodCommand      = composable.CreateCommand
	UpdateSalesFulfillmentMethodCommand      = composable.UpdateCommand
	DeleteSalesFulfillmentMethodCommand      = composable.DeleteCommand
	SetSalesFulfillmentMethodArchivedCommand = composable.SetArchivedCommand
	GetSalesFulfillmentMethodByIdQuery       = composable.GetByIdQuery
	SearchSalesFulfillmentMethodsQuery       = composable.SearchQuery
	SalesFulfillmentMethodExistsQuery        = composable.ExistsQuery
)

type (
	CreateSalesFulfillmentMethodResult      = composable.CreateResult
	UpdateSalesFulfillmentMethodResult      = composable.MutateResult
	DeleteSalesFulfillmentMethodResult      = composable.MutateResult
	SetSalesFulfillmentMethodArchivedResult = composable.MutateResult
	GetSalesFulfillmentMethodByIdResult     = composable.GetOneResult
	SearchSalesFulfillmentMethodsResult     = composable.SearchResult
	SalesFulfillmentMethodExistsResult      = composable.ExistsResult
)

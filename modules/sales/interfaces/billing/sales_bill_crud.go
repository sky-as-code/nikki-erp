package billing

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// The CRUD layers of the billing resources. The bill itself is writable and carries five actions;
// the rest are read-only, written by the bill's own operations.

// SalesBillRepository reads and writes bills.
type SalesBillRepository interface {
	composable.CrudRepository
}

// SalesBillDomainService is the CRUD of the resource. The bill's rules live in the operation
// services (split, merge, pay, settle) rather than in CRUD overrides, because each is a
// multi-row transaction guarded by a distributed lock rather than a check on one write.
type SalesBillDomainService interface {
	composable.CrudDomainService
}

// SalesBillApplicationService is the authorized surface the REST handler serves.
type SalesBillApplicationService interface {
	composable.CrudApplicationService

	// Split and Merge restructure what the customer owes; Pay and Settle move money. Each is a
	// power a grantor withholds independently of ordinary bill edits, so each carries its own
	// permission code.
	Split(ctx corectx.Context, cmd BillActionCommand) (*dyn.OpResult[any], error)

	// Merge is collection-level: it names several source bills rather than addressing one.
	Merge(ctx corectx.Context, cmd BillActionCommand) (*dyn.OpResult[any], error)

	Pay(ctx corectx.Context, cmd BillActionCommand) (*dyn.OpResult[any], error)
	Settle(ctx corectx.Context, cmd BillActionCommand) (*dyn.OpResult[any], error)

	// StartGatewayPayment shares the pay permission: handing the customer to a gateway is the same
	// power over the same money as recording a payment directly.
	StartGatewayPayment(ctx corectx.Context, cmd BillActionCommand) (*dyn.OpResult[any], error)
}

// BillActionCommand is the bound request of a custom bill action.
type BillActionCommand = composable.UpdateCommand

// The read-only billing resources. A line is written by the bill it belongs to, a relation by a
// split or merge, a payment by the pay and settle paths, a fulfillment request by confirmation.
// A client that could write them directly could contradict the bill they describe.

type SalesBillLineRepository interface {
	composable.CrudRepository
}

type SalesBillLineApplicationService interface {
	composable.CrudApplicationService
}

type SalesBillRelationRepository interface {
	composable.CrudRepository
}

type SalesBillRelationApplicationService interface {
	composable.CrudApplicationService
}

type SalesPaymentRepository interface {
	composable.CrudRepository
}

type SalesPaymentApplicationService interface {
	composable.CrudApplicationService
}

type SalesFulfillmentRequestRepository interface {
	composable.CrudRepository
}

type SalesFulfillmentRequestApplicationService interface {
	composable.CrudApplicationService
}

type SalesFulfillmentRequestLineRepository interface {
	composable.CrudRepository
}

type SalesFulfillmentRequestLineApplicationService interface {
	composable.CrudApplicationService
}

type (
	CreateSalesBillCommand      = composable.CreateCommand
	UpdateSalesBillCommand      = composable.UpdateCommand
	DeleteSalesBillCommand      = composable.DeleteCommand
	SetSalesBillArchivedCommand = composable.SetArchivedCommand
	GetSalesBillByIdQuery       = composable.GetByIdQuery
	SearchSalesBillsQuery       = composable.SearchQuery
	SalesBillExistsQuery        = composable.ExistsQuery
)

type (
	CreateSalesBillResult      = composable.CreateResult
	UpdateSalesBillResult      = composable.MutateResult
	DeleteSalesBillResult      = composable.MutateResult
	SetSalesBillArchivedResult = composable.MutateResult
	GetSalesBillByIdResult     = composable.GetOneResult
	SearchSalesBillsResult     = composable.SearchResult
	SalesBillExistsResult      = composable.ExistsResult
)

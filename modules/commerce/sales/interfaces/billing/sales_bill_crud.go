package billing

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// The CRUD layers of the billing resources. The bill itself is writable and carries five actions;
// the rest are read-only, written by the bill's own operations.

// SalesBillRepository reads and writes bills.
type SalesBillRepository interface {
	dynamicresource.CrudRepository
}

// SalesBillDomainService is the CRUD of the resource. The bill's rules live in the operation
// services (split, merge, pay, settle) rather than in CRUD overrides, because each is a
// multi-row transaction guarded by a distributed lock rather than a check on one write.
type SalesBillDomainService interface {
	dynamicresource.CrudDomainService
}

// SalesBillApplicationService is the authorized surface the REST handler serves.
type SalesBillApplicationService interface {
	dynamicresource.CrudApplicationService

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
type BillActionCommand = dynamicresource.UpdateCommand

// The read-only billing resources. A line is written by the bill it belongs to, a relation by a
// split or merge, a payment by the pay and settle paths, a fulfillment request by confirmation.
// A client that could write them directly could contradict the bill they describe.

type SalesBillLineRepository interface {
	dynamicresource.CrudRepository
}

type SalesBillLineApplicationService interface {
	dynamicresource.CrudApplicationService
}

type SalesBillRelationRepository interface {
	dynamicresource.CrudRepository
}

type SalesBillRelationApplicationService interface {
	dynamicresource.CrudApplicationService
}

type SalesPaymentRepository interface {
	dynamicresource.CrudRepository
}

type SalesPaymentApplicationService interface {
	dynamicresource.CrudApplicationService
}

type SalesFulfillmentRequestRepository interface {
	dynamicresource.CrudRepository
}

type SalesFulfillmentRequestApplicationService interface {
	dynamicresource.CrudApplicationService
}

type SalesFulfillmentRequestLineRepository interface {
	dynamicresource.CrudRepository
}

type SalesFulfillmentRequestLineApplicationService interface {
	dynamicresource.CrudApplicationService
}

type (
	CreateSalesBillCommand      = dynamicresource.CreateCommand
	UpdateSalesBillCommand      = dynamicresource.UpdateCommand
	DeleteSalesBillCommand      = dynamicresource.DeleteCommand
	SetSalesBillArchivedCommand = dynamicresource.SetArchivedCommand
	GetSalesBillByIdQuery       = dynamicresource.GetByIdQuery
	SearchSalesBillsQuery       = dynamicresource.SearchQuery
	SalesBillExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateSalesBillResult      = dynamicresource.CreateResult
	UpdateSalesBillResult      = dynamicresource.MutateResult
	DeleteSalesBillResult      = dynamicresource.MutateResult
	SetSalesBillArchivedResult = dynamicresource.MutateResult
	GetSalesBillByIdResult     = dynamicresource.GetOneResult
	SearchSalesBillsResult     = dynamicresource.SearchResult
	SalesBillExistsResult      = dynamicresource.ExistsResult
)

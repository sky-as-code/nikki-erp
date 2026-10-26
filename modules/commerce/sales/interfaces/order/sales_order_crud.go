package order

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// The CRUD layers of the order resources, beside the cross-module SalesOrderExtService that this
// package already declares. They are separate surfaces on purpose: the ext service is the in-process
// port another module calls, these are the resource's own authorized CRUD.

// SalesOrderRepository reads and writes sales orders.
type SalesOrderRepository interface {
	composable.CrudRepository
}

// SalesOrderDomainService adds the rules built-in CRUD cannot express. The framework declares no
// CHECK constraints, so this is the single enforcement point: a write that bypasses this service
// bypasses the invariant entirely.
type SalesOrderDomainService interface {
	composable.CrudDomainService

	// AssertEditable refuses a change to an order past draft. Confirmation is the line: after it the
	// numbers are what the business promised the customer.
	AssertEditable(ctx corectx.Context, orderId string) (*dyn.OpResult[dyn.MutateResultData], error)
}

// SalesOrderApplicationService is the authorized surface the REST handler serves. Each method
// asserts its permission before touching the domain: the composable RouteDefinition carries no
// permission field, so this layer is where authorization happens.
type SalesOrderApplicationService interface {
	composable.CrudApplicationService

	// Collection-level: there is no record to place in an org yet.
	CreateOrder(ctx corectx.Context, cmd OrderActionCommand) (*dyn.OpResult[any], error)

	Reprice(ctx corectx.Context, cmd OrderActionCommand) (*dyn.OpResult[any], error)
	Confirm(ctx corectx.Context, cmd OrderActionCommand) (*dyn.OpResult[any], error)
	Cancel(ctx corectx.Context, cmd OrderActionCommand) (*dyn.OpResult[any], error)
	ApplyVoucher(ctx corectx.Context, cmd OrderActionCommand) (*dyn.OpResult[any], error)
	ExplainPrice(ctx corectx.Context, query OrderActionCommand) (*dyn.OpResult[any], error)
	GrantManualDiscount(ctx corectx.Context, cmd OrderActionCommand) (*dyn.OpResult[any], error)
	RevokeManualDiscount(ctx corectx.Context, cmd OrderActionCommand) (*dyn.OpResult[any], error)

	AssignParties(ctx corectx.Context, cmd OrderActionCommand) (*dyn.OpResult[any], error)
	AssignSoldToParty(ctx corectx.Context, cmd OrderActionCommand) (*dyn.OpResult[any], error)
	AssignBillToParty(ctx corectx.Context, cmd OrderActionCommand) (*dyn.OpResult[any], error)
	AssignPayerParty(ctx corectx.Context, cmd OrderActionCommand) (*dyn.OpResult[any], error)

	// The fulfillment views. SearchFulfillmentTargets is collection-level: it asks which outlets
	// could satisfy a basket, naming no order.
	ListFulfillments(ctx corectx.Context, query OrderActionCommand) (*dyn.OpResult[any], error)
	SearchFulfillmentTargets(ctx corectx.Context, query OrderActionCommand) (*dyn.OpResult[any], error)

	// The refund views.
	CreateRefunds(ctx corectx.Context, cmd OrderActionCommand) (*dyn.OpResult[any], error)
	ViewRefunds(ctx corectx.Context, query OrderActionCommand) (*dyn.OpResult[any], error)

	// ListBills answers what this order owes and owed, superseded bills included, for a client that
	// lost a confirm response and for reconciliation after a split or a merge.
	ListBills(ctx corectx.Context, query OrderActionCommand) (*dyn.OpResult[any], error)
}

// OrderActionCommand is the bound request of a custom order action: the field map the REST engine
// produced, with the path id merged in. One alias for all of them because the actions differ in
// which fields they read, not in the shape they arrive as.
type OrderActionCommand = composable.UpdateCommand

// SalesOrderLineRepository reads and writes order lines.
type SalesOrderLineRepository interface {
	composable.CrudRepository
}

// SalesOrderLineDomainService carries the quantity invariant and the snapshot immutability rule.
type SalesOrderLineDomainService interface {
	composable.CrudDomainService
}

// SalesOrderLineAllocationRepository reads the persisted location split of an order line.
type SalesOrderLineAllocationRepository interface {
	composable.CrudRepository
}

// SalesOrderLineAllocationDomainService guards the allocation's single-row invariants. The
// cross-row check that allocations add up to the line quantity belongs to the create-order flow,
// where the whole allocation set is available.
type SalesOrderLineAllocationDomainService interface {
	composable.CrudDomainService
}

// SalesOrderLineApplicationService is the authorized surface the REST handler serves.
type SalesOrderLineApplicationService interface {
	composable.CrudApplicationService
}

// The read-only order resources. They are written by the order's own operations -- a component by
// repricing, an adjustment by a discount, an event by every transition -- so a client that could
// write them directly could contradict the order they describe.

type SalesOrderLineComponentRepository interface {
	composable.CrudRepository
}

type SalesOrderLineComponentApplicationService interface {
	composable.CrudApplicationService
}

type SalesOrderAdjustmentRepository interface {
	composable.CrudRepository
}

type SalesOrderAdjustmentApplicationService interface {
	composable.CrudApplicationService
}

type SalesOrderEventRepository interface {
	composable.CrudRepository
}

type SalesOrderEventApplicationService interface {
	composable.CrudApplicationService
}

type (
	CreateSalesOrderCrudCommand  = composable.CreateCommand
	UpdateSalesOrderCrudCommand  = composable.UpdateCommand
	DeleteSalesOrderCrudCommand  = composable.DeleteCommand
	SetSalesOrderArchivedCommand = composable.SetArchivedCommand
	GetSalesOrderByIdQuery       = composable.GetByIdQuery
	SearchSalesOrdersQuery       = composable.SearchQuery
	SalesOrderExistsQuery        = composable.ExistsQuery
)

type (
	CreateSalesOrderCrudResult  = composable.CreateResult
	UpdateSalesOrderCrudResult  = composable.MutateResult
	DeleteSalesOrderCrudResult  = composable.MutateResult
	SetSalesOrderArchivedResult = composable.MutateResult
	GetSalesOrderByIdResult     = composable.GetOneResult
	SearchSalesOrdersResult     = composable.SearchResult
	SalesOrderExistsResult      = composable.ExistsResult
)

type (
	CreateSalesOrderLineCommand      = composable.CreateCommand
	UpdateSalesOrderLineCommand      = composable.UpdateCommand
	DeleteSalesOrderLineCommand      = composable.DeleteCommand
	SetSalesOrderLineArchivedCommand = composable.SetArchivedCommand
	GetSalesOrderLineByIdQuery       = composable.GetByIdQuery
	SearchSalesOrderLinesQuery       = composable.SearchQuery
	SalesOrderLineExistsQuery        = composable.ExistsQuery
)

type (
	CreateSalesOrderLineResult      = composable.CreateResult
	UpdateSalesOrderLineResult      = composable.MutateResult
	DeleteSalesOrderLineResult      = composable.MutateResult
	SetSalesOrderLineArchivedResult = composable.MutateResult
	GetSalesOrderLineByIdResult     = composable.GetOneResult
	SearchSalesOrderLinesResult     = composable.SearchResult
	SalesOrderLineExistsResult      = composable.ExistsResult
)

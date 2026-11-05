package order

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// The CRUD layers of the order resources, beside the cross-module SalesOrderExtService that this
// package already declares. They are separate surfaces on purpose: the ext service is the in-process
// port another module calls, these are the resource's own authorized CRUD.

// SalesOrderRepository reads and writes sales orders.
type SalesOrderRepository interface {
	dynamicresource.CrudRepository
}

// SalesOrderDomainService adds the rules built-in CRUD cannot express. The framework declares no
// CHECK constraints, so this is the single enforcement point: a write that bypasses this service
// bypasses the invariant entirely.
type SalesOrderDomainService interface {
	dynamicresource.CrudDomainService

	// AssertEditable refuses a change to an order past draft. Confirmation is the line: after it the
	// numbers are what the business promised the customer.
	AssertEditable(ctx corectx.Context, orderId string) (*dyn.OpResult[dyn.MutateResultData], error)
}

// SalesOrderApplicationService is the authorized surface the REST handler serves. Each method
// asserts its permission before touching the domain: the composable RouteDefinition carries no
// permission field, so this layer is where authorization happens.
type SalesOrderApplicationService interface {
	dynamicresource.CrudApplicationService

	// Collection-level: there is no record to place in an org yet.
	CreateOrder(ctx corectx.Context, cmd OrderActionCommand) (*dyn.OpResult[any], error)

	Reprice(ctx corectx.Context, cmd OrderActionCommand) (*dyn.OpResult[any], error)
	Confirm(ctx corectx.Context, cmd OrderActionCommand) (*dyn.OpResult[any], error)
	Cancel(ctx corectx.Context, cmd OrderActionCommand) (*dyn.OpResult[any], error)
	ApplyVoucher(ctx corectx.Context, cmd OrderActionCommand) (*dyn.OpResult[any], error)
	// The draft promotion actions answer the PromotionQuote; none consumes usage (BR-V07).
	RemoveVoucher(ctx corectx.Context, cmd OrderActionCommand) (*dyn.OpResult[any], error)
	SelectPromotionReward(ctx corectx.Context, cmd OrderActionCommand) (*dyn.OpResult[any], error)
	PreviewPromotions(ctx corectx.Context, cmd OrderActionCommand) (*dyn.OpResult[any], error)
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
type OrderActionCommand = dynamicresource.UpdateCommand

// SalesOrderLineRepository reads and writes order lines.
type SalesOrderLineRepository interface {
	dynamicresource.CrudRepository
}

// SalesOrderLineDomainService carries the quantity invariant and the snapshot immutability rule.
type SalesOrderLineDomainService interface {
	dynamicresource.CrudDomainService
}

// SalesOrderLineAllocationRepository reads the persisted location split of an order line.
type SalesOrderLineAllocationRepository interface {
	dynamicresource.CrudRepository
}

// SalesOrderLineAllocationDomainService guards the allocation's single-row invariants. The
// cross-row check that allocations add up to the line quantity belongs to the create-order flow,
// where the whole allocation set is available.
type SalesOrderLineAllocationDomainService interface {
	dynamicresource.CrudDomainService
}

// SalesOrderLineApplicationService is the authorized surface the REST handler serves.
type SalesOrderLineApplicationService interface {
	dynamicresource.CrudApplicationService
}

// The read-only order resources. They are written by the order's own operations -- a component by
// repricing, an adjustment by a discount, an event by every transition -- so a client that could
// write them directly could contradict the order they describe.

type SalesOrderLineComponentRepository interface {
	dynamicresource.CrudRepository
}

type SalesOrderLineComponentApplicationService interface {
	dynamicresource.CrudApplicationService
}

type SalesOrderAdjustmentRepository interface {
	dynamicresource.CrudRepository
}

type SalesOrderAdjustmentApplicationService interface {
	dynamicresource.CrudApplicationService
}

type SalesOrderEventRepository interface {
	dynamicresource.CrudRepository
}

type SalesOrderEventApplicationService interface {
	dynamicresource.CrudApplicationService
}

type (
	CreateSalesOrderCrudCommand  = dynamicresource.CreateCommand
	UpdateSalesOrderCrudCommand  = dynamicresource.UpdateCommand
	DeleteSalesOrderCrudCommand  = dynamicresource.DeleteCommand
	SetSalesOrderArchivedCommand = dynamicresource.SetArchivedCommand
	GetSalesOrderByIdQuery       = dynamicresource.GetByIdQuery
	SearchSalesOrdersQuery       = dynamicresource.SearchQuery
	SalesOrderExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateSalesOrderCrudResult  = dynamicresource.CreateResult
	UpdateSalesOrderCrudResult  = dynamicresource.MutateResult
	DeleteSalesOrderCrudResult  = dynamicresource.MutateResult
	SetSalesOrderArchivedResult = dynamicresource.MutateResult
	GetSalesOrderByIdResult     = dynamicresource.GetOneResult
	SearchSalesOrdersResult     = dynamicresource.SearchResult
	SalesOrderExistsResult      = dynamicresource.ExistsResult
)

type (
	CreateSalesOrderLineCommand      = dynamicresource.CreateCommand
	UpdateSalesOrderLineCommand      = dynamicresource.UpdateCommand
	DeleteSalesOrderLineCommand      = dynamicresource.DeleteCommand
	SetSalesOrderLineArchivedCommand = dynamicresource.SetArchivedCommand
	GetSalesOrderLineByIdQuery       = dynamicresource.GetByIdQuery
	SearchSalesOrderLinesQuery       = dynamicresource.SearchQuery
	SalesOrderLineExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateSalesOrderLineResult      = dynamicresource.CreateResult
	UpdateSalesOrderLineResult      = dynamicresource.MutateResult
	DeleteSalesOrderLineResult      = dynamicresource.MutateResult
	SetSalesOrderLineArchivedResult = dynamicresource.MutateResult
	GetSalesOrderLineByIdResult     = dynamicresource.GetOneResult
	SearchSalesOrderLinesResult     = dynamicresource.SearchResult
	SalesOrderLineExistsResult      = dynamicresource.ExistsResult
)

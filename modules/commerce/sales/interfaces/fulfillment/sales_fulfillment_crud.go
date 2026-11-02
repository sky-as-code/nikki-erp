// Package fulfillment declares the layers of the resources that carry out a sale: the fulfillment
// a confirmed order owes, the attempts made against it, and the record of a target being moved.
package fulfillment

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// SalesOrderFulfillmentRepository reads and writes the fulfillments of an order.
type SalesOrderFulfillmentRepository interface {
	dynamicresource.CrudRepository
}

// SalesOrderFulfillmentDomainService is the CRUD of the resource. Its behaviour lives in the
// operation services (create, attempt, reassign, release), each a multi-row transaction rather
// than a check on a single write.
type SalesOrderFulfillmentDomainService interface {
	dynamicresource.CrudDomainService
}

// SalesOrderFulfillmentApplicationService is the authorized surface the REST handler serves.
//
// All three actions reuse the update permission: commanding a dispense, recording its result and
// moving a target all change what a delivery owes, which is the same power over the same sale.
type SalesOrderFulfillmentApplicationService interface {
	dynamicresource.CrudApplicationService

	CreateAttempt(ctx corectx.Context, cmd FulfillmentActionCommand) (*dyn.OpResult[any], error)
	ApplyAttemptResult(ctx corectx.Context, cmd FulfillmentActionCommand) (*dyn.OpResult[any], error)
	ReassignTarget(ctx corectx.Context, cmd FulfillmentActionCommand) (*dyn.OpResult[any], error)
}

// FulfillmentActionCommand is the bound request of a custom fulfillment action.
type FulfillmentActionCommand = dynamicresource.UpdateCommand

// The read-only fulfillment resources. An item is written by the fulfillment that owns it, an
// attempt and its items by the dispense loop, a target change by the reassignment that made it --
// each is a record of something that happened, which a client must not be able to rewrite.

type SalesOrderFulfillmentItemRepository interface {
	dynamicresource.CrudRepository
}

type SalesOrderFulfillmentItemApplicationService interface {
	dynamicresource.CrudApplicationService
}

type SalesFulfillmentAttemptRepository interface {
	dynamicresource.CrudRepository
}

type SalesFulfillmentAttemptApplicationService interface {
	dynamicresource.CrudApplicationService
}

type SalesFulfillmentAttemptItemRepository interface {
	dynamicresource.CrudRepository
}

type SalesFulfillmentAttemptItemApplicationService interface {
	dynamicresource.CrudApplicationService
}

type SalesFulfillmentTargetChangeRepository interface {
	dynamicresource.CrudRepository
}

type SalesFulfillmentTargetChangeApplicationService interface {
	dynamicresource.CrudApplicationService
}

type (
	CreateSalesOrderFulfillmentCommand      = dynamicresource.CreateCommand
	UpdateSalesOrderFulfillmentCommand      = dynamicresource.UpdateCommand
	DeleteSalesOrderFulfillmentCommand      = dynamicresource.DeleteCommand
	SetSalesOrderFulfillmentArchivedCommand = dynamicresource.SetArchivedCommand
	GetSalesOrderFulfillmentByIdQuery       = dynamicresource.GetByIdQuery
	SearchSalesOrderFulfillmentsQuery       = dynamicresource.SearchQuery
	SalesOrderFulfillmentExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateSalesOrderFulfillmentResult      = dynamicresource.CreateResult
	UpdateSalesOrderFulfillmentResult      = dynamicresource.MutateResult
	DeleteSalesOrderFulfillmentResult      = dynamicresource.MutateResult
	SetSalesOrderFulfillmentArchivedResult = dynamicresource.MutateResult
	GetSalesOrderFulfillmentByIdResult     = dynamicresource.GetOneResult
	SearchSalesOrderFulfillmentsResult     = dynamicresource.SearchResult
	SalesOrderFulfillmentExistsResult      = dynamicresource.ExistsResult
)

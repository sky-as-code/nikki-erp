// Package order declares the CRUD layers of the purchase order, its lines and its sourcing groups.
package order

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// ActionCommand is the bound request of a custom action: path, body and the resolved org.
type ActionCommand = composable.UpdateCommand

// ActionResult answers a custom action with the body the legacy engine action returned: a
// mutation summary, the created record, or the action's own report.
type ActionResult = dyn.OpResult[any]

// PurchaseOrderCrudRepository reads and writes purchase orders. An RFQ and a PO are one record at
// different statuses.
type PurchaseOrderCrudRepository interface {
	composable.CrudRepository
}

// PurchaseOrderCrudDomainService is the CRUD of the resource. Its Create stamps the fields a client
// may not choose, its Delete refuses an order that is not cancelled, and the concrete service
// carries the lifecycle operations.
type PurchaseOrderCrudDomainService interface {
	composable.CrudDomainService
}

// PurchaseOrderCrudApplicationService is the authorized surface the REST handler serves.
type PurchaseOrderCrudApplicationService interface {
	composable.CrudApplicationService

	// The lifecycle actions. Each asserts its own permission code: confirming or approving is a
	// materially different power from editing a description.
	ConfirmAction(ctx corectx.Context, cmd ActionCommand) (*ActionResult, error)
	ApproveAction(ctx corectx.Context, cmd ActionCommand) (*ActionResult, error)
	CancelAction(ctx corectx.Context, cmd ActionCommand) (*ActionResult, error)
	SendAction(ctx corectx.Context, cmd ActionCommand) (*ActionResult, error)
	LockAction(ctx corectx.Context, cmd ActionCommand) (*ActionResult, error)
	UnlockAction(ctx corectx.Context, cmd ActionCommand) (*ActionResult, error)
	AcknowledgeAction(ctx corectx.Context, cmd ActionCommand) (*ActionResult, error)
	RepriceAction(ctx corectx.Context, cmd ActionCommand) (*ActionResult, error)

	// MergeAction is collection-level: it names several orders in the body.
	MergeAction(ctx corectx.Context, cmd ActionCommand) (*ActionResult, error)

	// CreateAlternativeAction and DuplicateAction create an order, so they carry "create";
	// CompareAlternativesAction changes nothing, so it carries "read".
	CreateAlternativeAction(ctx corectx.Context, cmd ActionCommand) (*ActionResult, error)
	CompareAlternativesAction(ctx corectx.Context, cmd ActionCommand) (*ActionResult, error)
	DuplicateAction(ctx corectx.Context, cmd ActionCommand) (*ActionResult, error)
}

type (
	CreatePurchaseOrderCommand      = composable.CreateCommand
	UpdatePurchaseOrderCommand      = composable.UpdateCommand
	DeletePurchaseOrderCommand      = composable.DeleteCommand
	SetPurchaseOrderArchivedCommand = composable.SetArchivedCommand
	GetPurchaseOrderByIdQuery       = composable.GetByIdQuery
	SearchPurchaseOrdersQuery       = composable.SearchQuery
	PurchaseOrderExistsQuery        = composable.ExistsQuery
)

// PurchaseOrderLineCrudRepository reads and writes purchase order lines.
type PurchaseOrderLineCrudRepository interface {
	composable.CrudRepository
}

// PurchaseOrderLineCrudDomainService is the CRUD of the resource. Every write recomputes the order header totals in
// the same transaction.
type PurchaseOrderLineCrudDomainService interface {
	composable.CrudDomainService
}

// PurchaseOrderLineCrudApplicationService is the authorized surface the REST handler serves.
type PurchaseOrderLineCrudApplicationService interface {
	composable.CrudApplicationService
}

type (
	CreatePurchaseOrderLineCommand      = composable.CreateCommand
	UpdatePurchaseOrderLineCommand      = composable.UpdateCommand
	DeletePurchaseOrderLineCommand      = composable.DeleteCommand
	SetPurchaseOrderLineArchivedCommand = composable.SetArchivedCommand
	GetPurchaseOrderLineByIdQuery       = composable.GetByIdQuery
	SearchPurchaseOrderLinesQuery       = composable.SearchQuery
	PurchaseOrderLineExistsQuery        = composable.ExistsQuery
)

// SourcingGroupCrudRepository reads and writes sourcing groups.
type SourcingGroupCrudRepository interface {
	composable.CrudRepository
}

// SourcingGroupCrudDomainService is the CRUD of the resource. Create and delete are refused: a group is made by
// create_alternative and reaped when fewer than two alternatives remain.
type SourcingGroupCrudDomainService interface {
	composable.CrudDomainService
}

// SourcingGroupCrudApplicationService is the authorized surface the REST handler serves.
type SourcingGroupCrudApplicationService interface {
	composable.CrudApplicationService
}

type (
	CreateSourcingGroupCommand      = composable.CreateCommand
	UpdateSourcingGroupCommand      = composable.UpdateCommand
	DeleteSourcingGroupCommand      = composable.DeleteCommand
	SetSourcingGroupArchivedCommand = composable.SetArchivedCommand
	GetSourcingGroupByIdQuery       = composable.GetByIdQuery
	SearchSourcingGroupsQuery       = composable.SearchQuery
	SourcingGroupExistsQuery        = composable.ExistsQuery
)

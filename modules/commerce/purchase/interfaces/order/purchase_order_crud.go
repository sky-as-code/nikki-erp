// Package order declares the CRUD layers of the purchase order, its lines and its sourcing groups.
package order

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// ActionCommand is the bound request of a custom action: path, body and the resolved org.
type ActionCommand = dynamicresource.UpdateCommand

// ActionResult answers a custom action with the body the legacy engine action returned: a
// mutation summary, the created record, or the action's own report.
type ActionResult = dyn.OpResult[any]

// PurchaseOrderCrudRepository reads and writes purchase orders. An RFQ and a PO are one record at
// different statuses.
type PurchaseOrderCrudRepository interface {
	dynamicresource.CrudRepository
}

// PurchaseOrderCrudDomainService is the CRUD of the resource. Its Create stamps the fields a client
// may not choose, its Delete refuses an order that is not cancelled, and the concrete service
// carries the lifecycle operations.
type PurchaseOrderCrudDomainService interface {
	dynamicresource.CrudDomainService
}

// PurchaseOrderCrudApplicationService is the authorized surface the REST handler serves.
type PurchaseOrderCrudApplicationService interface {
	dynamicresource.CrudApplicationService

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
	CreatePurchaseOrderCommand      = dynamicresource.CreateCommand
	UpdatePurchaseOrderCommand      = dynamicresource.UpdateCommand
	DeletePurchaseOrderCommand      = dynamicresource.DeleteCommand
	SetPurchaseOrderArchivedCommand = dynamicresource.SetArchivedCommand
	GetPurchaseOrderByIdQuery       = dynamicresource.GetByIdQuery
	SearchPurchaseOrdersQuery       = dynamicresource.SearchQuery
	PurchaseOrderExistsQuery        = dynamicresource.ExistsQuery
)

// PurchaseOrderLineCrudRepository reads and writes purchase order lines.
type PurchaseOrderLineCrudRepository interface {
	dynamicresource.CrudRepository
}

// PurchaseOrderLineCrudDomainService is the CRUD of the resource. Every write recomputes the order header totals in
// the same transaction.
type PurchaseOrderLineCrudDomainService interface {
	dynamicresource.CrudDomainService
}

// PurchaseOrderLineCrudApplicationService is the authorized surface the REST handler serves.
type PurchaseOrderLineCrudApplicationService interface {
	dynamicresource.CrudApplicationService
}

type (
	CreatePurchaseOrderLineCommand      = dynamicresource.CreateCommand
	UpdatePurchaseOrderLineCommand      = dynamicresource.UpdateCommand
	DeletePurchaseOrderLineCommand      = dynamicresource.DeleteCommand
	SetPurchaseOrderLineArchivedCommand = dynamicresource.SetArchivedCommand
	GetPurchaseOrderLineByIdQuery       = dynamicresource.GetByIdQuery
	SearchPurchaseOrderLinesQuery       = dynamicresource.SearchQuery
	PurchaseOrderLineExistsQuery        = dynamicresource.ExistsQuery
)

// SourcingGroupCrudRepository reads and writes sourcing groups.
type SourcingGroupCrudRepository interface {
	dynamicresource.CrudRepository
}

// SourcingGroupCrudDomainService is the CRUD of the resource. Create and delete are refused: a group is made by
// create_alternative and reaped when fewer than two alternatives remain.
type SourcingGroupCrudDomainService interface {
	dynamicresource.CrudDomainService
}

// SourcingGroupCrudApplicationService is the authorized surface the REST handler serves.
type SourcingGroupCrudApplicationService interface {
	dynamicresource.CrudApplicationService
}

type (
	CreateSourcingGroupCommand      = dynamicresource.CreateCommand
	UpdateSourcingGroupCommand      = dynamicresource.UpdateCommand
	DeleteSourcingGroupCommand      = dynamicresource.DeleteCommand
	SetSourcingGroupArchivedCommand = dynamicresource.SetArchivedCommand
	GetSourcingGroupByIdQuery       = dynamicresource.GetByIdQuery
	SearchSourcingGroupsQuery       = dynamicresource.SearchQuery
	SourcingGroupExistsQuery        = dynamicresource.ExistsQuery
)

package order

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// OrderCrudRepository reads and writes payment orders.
type OrderCrudRepository interface {
	dynamicresource.CrudRepository
}

// OrderCrudDomainService is the CRUD of the resource. The money-moving operations are not CRUD and
// live on the OrderDomainService port instead.
type OrderCrudDomainService interface {
	dynamicresource.CrudDomainService
}

// OrderCrudApplicationService is the authorized surface the REST handler serves. Each action
// asserts its own permission code before delegating to the OrderDomainService port: being allowed
// to correct an order is not the same authority as being allowed to hand money back.
type OrderCrudApplicationService interface {
	dynamicresource.CrudApplicationService

	CreatePaymentAction(ctx corectx.Context, cmd OrderActionCommand) (*OrderActionResult, error)
	RefundAction(ctx corectx.Context, cmd OrderActionCommand) (*OrderActionResult, error)
	RemovePosOrdersAction(ctx corectx.Context, cmd OrderActionCommand) (*OrderActionResult, error)
}

// OrderActionCommand is the bound request of a custom order action: path, body and resolved org.
type OrderActionCommand = dynamicresource.UpdateCommand

// OrderActionResult answers a custom order action with the flat body the legacy actions returned.
type OrderActionResult = dyn.OpResult[map[string]any]

type (
	CreateOrderCommand      = dynamicresource.CreateCommand
	UpdateOrderCommand      = dynamicresource.UpdateCommand
	DeleteOrderCommand      = dynamicresource.DeleteCommand
	SetOrderArchivedCommand = dynamicresource.SetArchivedCommand
	GetOrderByIdQuery       = dynamicresource.GetByIdQuery
	SearchOrdersQuery       = dynamicresource.SearchQuery
	OrderExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateOrderResult      = dynamicresource.CreateResult
	UpdateOrderResult      = dynamicresource.MutateResult
	DeleteOrderResult      = dynamicresource.MutateResult
	SetOrderArchivedResult = dynamicresource.MutateResult
	GetOrderByIdResult     = dynamicresource.GetOneResult
	SearchOrdersResult     = dynamicresource.SearchResult
	OrderExistsResult      = dynamicresource.ExistsResult
)

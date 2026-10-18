package order

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// OrderCrudRepository reads and writes payment orders.
type OrderCrudRepository interface {
	composable.CrudRepository
}

// OrderCrudDomainService is the CRUD of the resource. The money-moving operations are not CRUD and
// live on the OrderDomainService port instead.
type OrderCrudDomainService interface {
	composable.CrudDomainService
}

// OrderCrudApplicationService is the authorized surface the REST handler serves. Each action
// asserts its own permission code before delegating to the OrderDomainService port: being allowed
// to correct an order is not the same authority as being allowed to hand money back.
type OrderCrudApplicationService interface {
	composable.CrudApplicationService

	CreatePaymentAction(ctx corectx.Context, cmd OrderActionCommand) (*OrderActionResult, error)
	RefundAction(ctx corectx.Context, cmd OrderActionCommand) (*OrderActionResult, error)
	RemovePosOrdersAction(ctx corectx.Context, cmd OrderActionCommand) (*OrderActionResult, error)
}

// OrderActionCommand is the bound request of a custom order action: path, body and resolved org.
type OrderActionCommand = composable.UpdateCommand

// OrderActionResult answers a custom order action with the flat body the legacy actions returned.
type OrderActionResult = dyn.OpResult[map[string]any]

type (
	CreateOrderCommand      = composable.CreateCommand
	UpdateOrderCommand      = composable.UpdateCommand
	DeleteOrderCommand      = composable.DeleteCommand
	SetOrderArchivedCommand = composable.SetArchivedCommand
	GetOrderByIdQuery       = composable.GetByIdQuery
	SearchOrdersQuery       = composable.SearchQuery
	OrderExistsQuery        = composable.ExistsQuery
)

type (
	CreateOrderResult      = composable.CreateResult
	UpdateOrderResult      = composable.MutateResult
	DeleteOrderResult      = composable.MutateResult
	SetOrderArchivedResult = composable.MutateResult
	GetOrderByIdResult     = composable.GetOneResult
	SearchOrdersResult     = composable.SearchResult
	OrderExistsResult      = composable.ExistsResult
)

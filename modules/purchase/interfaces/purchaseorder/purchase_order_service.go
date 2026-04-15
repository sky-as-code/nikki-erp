package purchaseorder

import corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

type PurchaseOrderService interface {
	CreatePurchaseOrder(ctx corectx.Context, cmd CreatePurchaseOrderCommand) (*CreatePurchaseOrderResult, error)
	DeletePurchaseOrder(ctx corectx.Context, cmd DeletePurchaseOrderCommand) (*DeletePurchaseOrderResult, error)
	PurchaseOrderExists(ctx corectx.Context, query PurchaseOrderExistsQuery) (*PurchaseOrderExistsResult, error)
	GetPurchaseOrder(ctx corectx.Context, query GetPurchaseOrderQuery) (*GetPurchaseOrderResult, error)
	SearchPurchaseOrders(ctx corectx.Context, query SearchPurchaseOrdersQuery) (*SearchPurchaseOrdersResult, error)
	SetPurchaseOrderIsArchived(ctx corectx.Context, cmd SetPurchaseOrderIsArchivedCommand) (*SetPurchaseOrderIsArchivedResult, error)
	UpdatePurchaseOrder(ctx corectx.Context, cmd UpdatePurchaseOrderCommand) (*UpdatePurchaseOrderResult, error)
}

package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	it "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/purchaseorder"
)

func NewPurchaseOrderApplicationServiceImpl(purchaseOrderSvc it.PurchaseOrderDomainService) it.PurchaseOrderAppService {
	return &PurchaseOrderApplicationServiceImpl{purchaseOrderSvc: purchaseOrderSvc}
}

type PurchaseOrderApplicationServiceImpl struct {
	purchaseOrderSvc it.PurchaseOrderDomainService
}

func (this *PurchaseOrderApplicationServiceImpl) CreatePurchaseOrder(ctx corectx.Context, cmd it.CreatePurchaseOrderCommand) (*it.CreatePurchaseOrderResult, error) {
	return this.purchaseOrderSvc.CreatePurchaseOrder(ctx, cmd)
}

func (this *PurchaseOrderApplicationServiceImpl) DeletePurchaseOrder(ctx corectx.Context, cmd it.DeletePurchaseOrderCommand) (*it.DeletePurchaseOrderResult, error) {
	return this.purchaseOrderSvc.DeletePurchaseOrder(ctx, cmd)
}

func (this *PurchaseOrderApplicationServiceImpl) PurchaseOrderExists(ctx corectx.Context, query it.PurchaseOrderExistsQuery) (*it.PurchaseOrderExistsResult, error) {
	return this.purchaseOrderSvc.PurchaseOrderExists(ctx, query)
}

func (this *PurchaseOrderApplicationServiceImpl) GetPurchaseOrder(ctx corectx.Context, query it.GetPurchaseOrderQuery) (*it.GetPurchaseOrderResult, error) {
	return this.purchaseOrderSvc.GetPurchaseOrder(ctx, query)
}

func (this *PurchaseOrderApplicationServiceImpl) SearchPurchaseOrders(ctx corectx.Context, query it.SearchPurchaseOrdersQuery) (*it.SearchPurchaseOrdersResult, error) {
	return this.purchaseOrderSvc.SearchPurchaseOrders(ctx, query)
}

func (this *PurchaseOrderApplicationServiceImpl) SetPurchaseOrderIsArchived(ctx corectx.Context, cmd it.SetPurchaseOrderIsArchivedCommand) (*it.SetPurchaseOrderIsArchivedResult, error) {
	return this.purchaseOrderSvc.SetPurchaseOrderIsArchived(ctx, cmd)
}

func (this *PurchaseOrderApplicationServiceImpl) UpdatePurchaseOrder(ctx corectx.Context, cmd it.UpdatePurchaseOrderCommand) (*it.UpdatePurchaseOrderResult, error) {
	return this.purchaseOrderSvc.UpdatePurchaseOrder(ctx, cmd)
}

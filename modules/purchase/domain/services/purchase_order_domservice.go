package services

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/purchaseorder"
)

func NewPurchaseOrderDomainServiceImpl(repo it.PurchaseOrderRepository) it.PurchaseOrderDomainService {
	return &PurchaseOrderDomainServiceImpl{repo: repo}
}

type PurchaseOrderDomainServiceImpl struct{ repo it.PurchaseOrderRepository }

func (this *PurchaseOrderDomainServiceImpl) CreatePurchaseOrder(ctx corectx.Context, cmd it.CreatePurchaseOrderCommand) (*it.CreatePurchaseOrderResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[models.PurchaseOrder, *models.PurchaseOrder]{
		Action: "create purchase order", BaseRepoGetter: this.repo, Data: cmd,
	})
}
func (this *PurchaseOrderDomainServiceImpl) DeletePurchaseOrder(ctx corectx.Context, cmd it.DeletePurchaseOrderCommand) (*it.DeletePurchaseOrderResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{Action: "delete purchase order", DbRepoGetter: this.repo, Cmd: dyn.DeleteOneCommand(cmd)})
}
func (this *PurchaseOrderDomainServiceImpl) PurchaseOrderExists(ctx corectx.Context, query it.PurchaseOrderExistsQuery) (*it.PurchaseOrderExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{Action: "check if purchase orders exist", DbRepoGetter: this.repo, Query: dyn.ExistsQuery(query)})
}
func (this *PurchaseOrderDomainServiceImpl) GetPurchaseOrder(ctx corectx.Context, query it.GetPurchaseOrderQuery) (*it.GetPurchaseOrderResult, error) {
	return corecrud.GetOne[models.PurchaseOrder](ctx, corecrud.GetOneParam{Action: "get purchase order", DbRepoGetter: this.repo, Query: dyn.GetOneQuery(query)})
}
func (this *PurchaseOrderDomainServiceImpl) SearchPurchaseOrders(ctx corectx.Context, query it.SearchPurchaseOrdersQuery) (*it.SearchPurchaseOrdersResult, error) {
	return corecrud.Search[models.PurchaseOrder](ctx, corecrud.SearchParam{Action: "search purchase orders", DbRepoGetter: this.repo, Query: dyn.SearchQuery(query)})
}
func (this *PurchaseOrderDomainServiceImpl) SetPurchaseOrderIsArchived(ctx corectx.Context, cmd it.SetPurchaseOrderIsArchivedCommand) (*it.SetPurchaseOrderIsArchivedResult, error) {
	return corecrud.SetIsArchived(ctx, this.repo, dyn.SetIsArchivedCommand(cmd))
}
func (this *PurchaseOrderDomainServiceImpl) UpdatePurchaseOrder(ctx corectx.Context, cmd it.UpdatePurchaseOrderCommand) (*it.UpdatePurchaseOrderResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[models.PurchaseOrder, *models.PurchaseOrder]{
		Action: "update purchase order", DbRepoGetter: this.repo, Data: cmd,
	})
}

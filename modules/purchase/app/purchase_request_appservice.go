package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	it "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/purchaserequest"
)

func NewPurchaseRequestApplicationServiceImpl(purchaseRequestSvc it.PurchaseRequestDomainService) it.PurchaseRequestAppService {
	return &PurchaseRequestApplicationServiceImpl{purchaseRequestSvc: purchaseRequestSvc}
}

type PurchaseRequestApplicationServiceImpl struct {
	purchaseRequestSvc it.PurchaseRequestDomainService
}

func (this *PurchaseRequestApplicationServiceImpl) CreatePurchaseRequest(ctx corectx.Context, cmd it.CreatePurchaseRequestCommand) (*it.CreatePurchaseRequestResult, error) {
	return this.purchaseRequestSvc.CreatePurchaseRequest(ctx, cmd)
}

func (this *PurchaseRequestApplicationServiceImpl) DeletePurchaseRequest(ctx corectx.Context, cmd it.DeletePurchaseRequestCommand) (*it.DeletePurchaseRequestResult, error) {
	return this.purchaseRequestSvc.DeletePurchaseRequest(ctx, cmd)
}

func (this *PurchaseRequestApplicationServiceImpl) PurchaseRequestExists(ctx corectx.Context, query it.PurchaseRequestExistsQuery) (*it.PurchaseRequestExistsResult, error) {
	return this.purchaseRequestSvc.PurchaseRequestExists(ctx, query)
}

func (this *PurchaseRequestApplicationServiceImpl) GetPurchaseRequest(ctx corectx.Context, query it.GetPurchaseRequestQuery) (*it.GetPurchaseRequestResult, error) {
	return this.purchaseRequestSvc.GetPurchaseRequest(ctx, query)
}

func (this *PurchaseRequestApplicationServiceImpl) SearchPurchaseRequests(ctx corectx.Context, query it.SearchPurchaseRequestsQuery) (*it.SearchPurchaseRequestsResult, error) {
	return this.purchaseRequestSvc.SearchPurchaseRequests(ctx, query)
}

func (this *PurchaseRequestApplicationServiceImpl) SetPurchaseRequestIsArchived(ctx corectx.Context, cmd it.SetPurchaseRequestIsArchivedCommand) (*it.SetPurchaseRequestIsArchivedResult, error) {
	return this.purchaseRequestSvc.SetPurchaseRequestIsArchived(ctx, cmd)
}

func (this *PurchaseRequestApplicationServiceImpl) UpdatePurchaseRequest(ctx corectx.Context, cmd it.UpdatePurchaseRequestCommand) (*it.UpdatePurchaseRequestResult, error) {
	return this.purchaseRequestSvc.UpdatePurchaseRequest(ctx, cmd)
}

func (this *PurchaseRequestApplicationServiceImpl) SubmitPurchaseRequestForApproval(ctx corectx.Context, cmd it.SubmitPurchaseRequestForApprovalCommand) (*it.SubmitPurchaseRequestForApprovalResult, error) {
	return this.purchaseRequestSvc.SubmitPurchaseRequestForApproval(ctx, cmd)
}

func (this *PurchaseRequestApplicationServiceImpl) ApprovePurchaseRequest(ctx corectx.Context, cmd it.ApprovePurchaseRequestCommand) (*it.ApprovePurchaseRequestResult, error) {
	return this.purchaseRequestSvc.ApprovePurchaseRequest(ctx, cmd)
}

func (this *PurchaseRequestApplicationServiceImpl) RejectPurchaseRequest(ctx corectx.Context, cmd it.RejectPurchaseRequestCommand) (*it.RejectPurchaseRequestResult, error) {
	return this.purchaseRequestSvc.RejectPurchaseRequest(ctx, cmd)
}

func (this *PurchaseRequestApplicationServiceImpl) CancelPurchaseRequest(ctx corectx.Context, cmd it.CancelPurchaseRequestCommand) (*it.CancelPurchaseRequestResult, error) {
	return this.purchaseRequestSvc.CancelPurchaseRequest(ctx, cmd)
}

func (this *PurchaseRequestApplicationServiceImpl) MarkPurchaseRequestPriority(ctx corectx.Context, cmd it.MarkPurchaseRequestPriorityCommand) (*it.MarkPurchaseRequestPriorityResult, error) {
	return this.purchaseRequestSvc.MarkPurchaseRequestPriority(ctx, cmd)
}

func (this *PurchaseRequestApplicationServiceImpl) ConvertPurchaseRequestToRfq(ctx corectx.Context, cmd it.ConvertPurchaseRequestToRfqCommand) (*it.ConvertPurchaseRequestToRfqResult, error) {
	return this.purchaseRequestSvc.ConvertPurchaseRequestToRfq(ctx, cmd)
}

func (this *PurchaseRequestApplicationServiceImpl) ConvertPurchaseRequestToPo(ctx corectx.Context, cmd it.ConvertPurchaseRequestToPoCommand) (*it.ConvertPurchaseRequestToPoResult, error) {
	return this.purchaseRequestSvc.ConvertPurchaseRequestToPo(ctx, cmd)
}

func (this *PurchaseRequestApplicationServiceImpl) ConsolidatePurchaseRequests(ctx corectx.Context, cmd it.ConsolidatePurchaseRequestsCommand) (*it.ConsolidatePurchaseRequestsResult, error) {
	return this.purchaseRequestSvc.ConsolidatePurchaseRequests(ctx, cmd)
}

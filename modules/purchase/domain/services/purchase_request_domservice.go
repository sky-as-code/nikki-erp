package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/purchaserequest"
)

func NewPurchaseRequestDomainServiceImpl(repo it.PurchaseRequestRepository) it.PurchaseRequestDomainService {
	return &PurchaseRequestDomainServiceImpl{repo: repo}
}

type PurchaseRequestDomainServiceImpl struct {
	repo it.PurchaseRequestRepository
}

func (this *PurchaseRequestDomainServiceImpl) CreatePurchaseRequest(
	ctx corectx.Context, cmd it.CreatePurchaseRequestCommand,
) (*it.CreatePurchaseRequestResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[models.PurchaseRequest, *models.PurchaseRequest]{
		Action:         "create purchase request",
		BaseRepoGetter: this.repo,
		Data:           cmd,
	})
}

func (this *PurchaseRequestDomainServiceImpl) DeletePurchaseRequest(
	ctx corectx.Context, cmd it.DeletePurchaseRequestCommand,
) (*it.DeletePurchaseRequestResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:       "delete purchase request",
		DbRepoGetter: this.repo,
		Cmd:          dyn.DeleteOneCommand(cmd),
	})
}

func (this *PurchaseRequestDomainServiceImpl) PurchaseRequestExists(
	ctx corectx.Context, query it.PurchaseRequestExistsQuery,
) (*it.PurchaseRequestExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "check if purchase requests exist",
		DbRepoGetter: this.repo,
		Query:        dyn.ExistsQuery(query),
	})
}

func (this *PurchaseRequestDomainServiceImpl) GetPurchaseRequest(
	ctx corectx.Context, query it.GetPurchaseRequestQuery,
) (*it.GetPurchaseRequestResult, error) {
	return corecrud.GetOne[models.PurchaseRequest](ctx, corecrud.GetOneParam{
		Action:       "get purchase request",
		DbRepoGetter: this.repo,
		Query:        dyn.GetOneQuery(query),
	})
}

func (this *PurchaseRequestDomainServiceImpl) SearchPurchaseRequests(
	ctx corectx.Context, query it.SearchPurchaseRequestsQuery,
) (*it.SearchPurchaseRequestsResult, error) {
	return corecrud.Search[models.PurchaseRequest](ctx, corecrud.SearchParam{
		Action:       "search purchase requests",
		DbRepoGetter: this.repo,
		Query:        dyn.SearchQuery(query),
	})
}

func (this *PurchaseRequestDomainServiceImpl) SetPurchaseRequestIsArchived(
	ctx corectx.Context, cmd it.SetPurchaseRequestIsArchivedCommand,
) (*it.SetPurchaseRequestIsArchivedResult, error) {
	return corecrud.SetIsArchived(ctx, this.repo, dyn.SetIsArchivedCommand(cmd))
}

func (this *PurchaseRequestDomainServiceImpl) UpdatePurchaseRequest(
	ctx corectx.Context, cmd it.UpdatePurchaseRequestCommand,
) (*it.UpdatePurchaseRequestResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[models.PurchaseRequest, *models.PurchaseRequest]{
		Action:       "update purchase request",
		DbRepoGetter: this.repo,
		Data:         cmd,
	})
}

func (this *PurchaseRequestDomainServiceImpl) SubmitPurchaseRequestForApproval(
	ctx corectx.Context, cmd it.SubmitPurchaseRequestForApprovalCommand,
) (*it.SubmitPurchaseRequestForApprovalResult, error) {
	return this.updateByAction(ctx, "submit purchase request for approval", cmd.Id, model.Etag(cmd.Etag), map[string]any{
		models.PurchaseRequestFieldStatus: models.PurchaseRequestStatusPendingApproval,
	})
}

func (this *PurchaseRequestDomainServiceImpl) ApprovePurchaseRequest(
	ctx corectx.Context, cmd it.ApprovePurchaseRequestCommand,
) (*it.ApprovePurchaseRequestResult, error) {
	return this.updateByAction(ctx, "approve purchase request", cmd.Id, model.Etag(cmd.Etag), map[string]any{
		models.PurchaseRequestFieldStatus: models.PurchaseRequestStatusApproved,
	})
}

func (this *PurchaseRequestDomainServiceImpl) RejectPurchaseRequest(
	ctx corectx.Context, cmd it.RejectPurchaseRequestCommand,
) (*it.RejectPurchaseRequestResult, error) {
	return this.updateByAction(ctx, "reject purchase request", cmd.Id, model.Etag(cmd.Etag), map[string]any{
		models.PurchaseRequestFieldStatus: models.PurchaseRequestStatusRejected,
	})
}

func (this *PurchaseRequestDomainServiceImpl) CancelPurchaseRequest(
	ctx corectx.Context, cmd it.CancelPurchaseRequestCommand,
) (*it.CancelPurchaseRequestResult, error) {
	return this.updateByAction(ctx, "cancel purchase request", cmd.Id, model.Etag(cmd.Etag), map[string]any{
		models.PurchaseRequestFieldStatus: models.PurchaseRequestStatusCancelled,
	})
}

func (this *PurchaseRequestDomainServiceImpl) MarkPurchaseRequestPriority(
	ctx corectx.Context, cmd it.MarkPurchaseRequestPriorityCommand,
) (*it.MarkPurchaseRequestPriorityResult, error) {
	return this.updateByAction(ctx, "mark purchase request priority", cmd.Id, model.Etag(cmd.Etag), map[string]any{
		models.PurchaseRequestFieldPriority: cmd.Priority,
	})
}

func (this *PurchaseRequestDomainServiceImpl) ConvertPurchaseRequestToRfq(
	ctx corectx.Context, cmd it.ConvertPurchaseRequestToRfqCommand,
) (*it.ConvertPurchaseRequestToRfqResult, error) {
	return this.updateByAction(ctx, "convert purchase request to rfq", cmd.Id, model.Etag(cmd.Etag), map[string]any{
		models.PurchaseRequestFieldConversionType: "rfq",
		models.PurchaseRequestFieldStatus:         models.PurchaseRequestStatusConvertedToRfq,
	})
}

func (this *PurchaseRequestDomainServiceImpl) ConvertPurchaseRequestToPo(
	ctx corectx.Context, cmd it.ConvertPurchaseRequestToPoCommand,
) (*it.ConvertPurchaseRequestToPoResult, error) {
	return this.updateByAction(ctx, "convert purchase request to po", cmd.Id, model.Etag(cmd.Etag), map[string]any{
		models.PurchaseRequestFieldConversionType: "po",
		models.PurchaseRequestFieldStatus:         models.PurchaseRequestStatusConvertedToPo,
	})
}

func (this *PurchaseRequestDomainServiceImpl) ConsolidatePurchaseRequests(
	ctx corectx.Context, cmd it.ConsolidatePurchaseRequestsCommand,
) (*it.ConsolidatePurchaseRequestsResult, error) {
	if len(cmd.SourcePurchaseRequestIds) == 0 {
		cErrs := ft.NewClientErrors()
		cErrs.Append(*ft.NewValidationError("source_purchase_request_ids", "required", "source purchase request ids are required"))
		return &it.ConsolidatePurchaseRequestsResult{
			ClientErrors: *cErrs,
		}, nil
	}

	return this.updateByAction(
		ctx,
		"consolidate purchase requests",
		cmd.SourcePurchaseRequestIds[0],
		model.Etag(cmd.Etag),
		map[string]any{
			models.PurchaseRequestFieldConversionType: "po",
			models.PurchaseRequestFieldStatus:         models.PurchaseRequestStatusConvertedToPo,
		},
	)
}

func (this *PurchaseRequestDomainServiceImpl) updateByAction(
	ctx corectx.Context, action string, id model.Id, etag model.Etag, fields map[string]any,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	data := dmodel.DynamicFields{
		models.PurchaseRequestFieldId: id,
		basemodel.FieldEtag:           etag,
	}
	for key, val := range fields {
		data.SetAny(key, val)
	}
	return corecrud.UpdateRegardless(ctx, corecrud.UpdateRegardlessParam{
		Action:       action,
		DbRepoGetter: this.repo,
		Data:         data,
	})
}

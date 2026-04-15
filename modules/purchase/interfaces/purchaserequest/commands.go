package purchaserequest

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain"
)

func init() {
	var req cqrs.Request
	req = (*CreatePurchaseRequestCommand)(nil)
	req = (*DeletePurchaseRequestCommand)(nil)
	req = (*GetPurchaseRequestQuery)(nil)
	req = (*SearchPurchaseRequestsQuery)(nil)
	req = (*UpdatePurchaseRequestCommand)(nil)
	req = (*SetPurchaseRequestIsArchivedCommand)(nil)
	req = (*PurchaseRequestExistsQuery)(nil)
	req = (*SubmitPurchaseRequestForApprovalCommand)(nil)
	req = (*ApprovePurchaseRequestCommand)(nil)
	req = (*RejectPurchaseRequestCommand)(nil)
	req = (*CancelPurchaseRequestCommand)(nil)
	req = (*MarkPurchaseRequestPriorityCommand)(nil)
	req = (*ConvertPurchaseRequestToRfqCommand)(nil)
	req = (*ConvertPurchaseRequestToPoCommand)(nil)
	req = (*ConsolidatePurchaseRequestsCommand)(nil)
	util.Unused(req)
}

var createPurchaseRequestCommandType = cqrs.RequestType{Module: "purchase", Submodule: "purchaserequest", Action: "create"}

type CreatePurchaseRequestCommand struct{ domain.PurchaseRequest }

func (CreatePurchaseRequestCommand) CqrsRequestType() cqrs.RequestType {
	return createPurchaseRequestCommandType
}
func (CreatePurchaseRequestCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.PurchaseRequestSchemaName)
}

type CreatePurchaseRequestResult = dyn.OpResult[domain.PurchaseRequest]

var updatePurchaseRequestCommandType = cqrs.RequestType{Module: "purchase", Submodule: "purchaserequest", Action: "update"}

type UpdatePurchaseRequestCommand struct{ domain.PurchaseRequest }

func (UpdatePurchaseRequestCommand) CqrsRequestType() cqrs.RequestType {
	return updatePurchaseRequestCommandType
}
func (UpdatePurchaseRequestCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.PurchaseRequestSchemaName)
}

type UpdatePurchaseRequestResult = dyn.OpResult[dyn.MutateResultData]

var deletePurchaseRequestCommandType = cqrs.RequestType{Module: "purchase", Submodule: "purchaserequest", Action: "delete"}

type DeletePurchaseRequestCommand dyn.DeleteOneCommand

func (DeletePurchaseRequestCommand) CqrsRequestType() cqrs.RequestType {
	return deletePurchaseRequestCommandType
}

type DeletePurchaseRequestResult = dyn.OpResult[dyn.MutateResultData]

var getPurchaseRequestQueryType = cqrs.RequestType{Module: "purchase", Submodule: "purchaserequest", Action: "get"}

type GetPurchaseRequestQuery dyn.GetOneQuery

func (GetPurchaseRequestQuery) CqrsRequestType() cqrs.RequestType { return getPurchaseRequestQueryType }

type GetPurchaseRequestResult = dyn.OpResult[domain.PurchaseRequest]

var searchPurchaseRequestsQueryType = cqrs.RequestType{Module: "purchase", Submodule: "purchaserequest", Action: "search"}

type SearchPurchaseRequestsQuery dyn.SearchQuery

func (SearchPurchaseRequestsQuery) CqrsRequestType() cqrs.RequestType {
	return searchPurchaseRequestsQueryType
}

type SearchPurchaseRequestsResultData = dyn.PagedResultData[domain.PurchaseRequest]
type SearchPurchaseRequestsResult = dyn.OpResult[SearchPurchaseRequestsResultData]

var setPurchaseRequestIsArchivedCommandType = cqrs.RequestType{
	Module: "purchase", Submodule: "purchaserequest", Action: "set_archived",
}

type SetPurchaseRequestIsArchivedCommand dyn.SetIsArchivedCommand

func (SetPurchaseRequestIsArchivedCommand) CqrsRequestType() cqrs.RequestType {
	return setPurchaseRequestIsArchivedCommandType
}

type SetPurchaseRequestIsArchivedResult = dyn.OpResult[dyn.MutateResultData]

var purchaseRequestExistsQueryType = cqrs.RequestType{
	Module: "purchase", Submodule: "purchaserequest", Action: "exists",
}

type PurchaseRequestExistsQuery dyn.ExistsQuery

func (PurchaseRequestExistsQuery) CqrsRequestType() cqrs.RequestType {
	return purchaseRequestExistsQueryType
}

type PurchaseRequestExistsResult = dyn.OpResult[dyn.ExistsResultData]

type submitPurchaseRequestForApprovalCommand struct {
	Id   model.Id `json:"id" param:"id"`
	Etag string   `json:"etag"`
}

var submitPurchaseRequestForApprovalCommandType = cqrs.RequestType{
	Module: "purchase", Submodule: "purchaserequest", Action: "submit_for_approval",
}

type SubmitPurchaseRequestForApprovalCommand submitPurchaseRequestForApprovalCommand

func (SubmitPurchaseRequestForApprovalCommand) CqrsRequestType() cqrs.RequestType {
	return submitPurchaseRequestForApprovalCommandType
}

type SubmitPurchaseRequestForApprovalResult = dyn.OpResult[dyn.MutateResultData]

var approvePurchaseRequestCommandType = cqrs.RequestType{
	Module: "purchase", Submodule: "purchaserequest", Action: "approve",
}

type ApprovePurchaseRequestCommand submitPurchaseRequestForApprovalCommand

func (ApprovePurchaseRequestCommand) CqrsRequestType() cqrs.RequestType {
	return approvePurchaseRequestCommandType
}

type ApprovePurchaseRequestResult = dyn.OpResult[dyn.MutateResultData]

var rejectPurchaseRequestCommandType = cqrs.RequestType{
	Module: "purchase", Submodule: "purchaserequest", Action: "reject",
}

type RejectPurchaseRequestCommand submitPurchaseRequestForApprovalCommand

func (RejectPurchaseRequestCommand) CqrsRequestType() cqrs.RequestType {
	return rejectPurchaseRequestCommandType
}

type RejectPurchaseRequestResult = dyn.OpResult[dyn.MutateResultData]

var cancelPurchaseRequestCommandType = cqrs.RequestType{
	Module: "purchase", Submodule: "purchaserequest", Action: "cancel",
}

type CancelPurchaseRequestCommand submitPurchaseRequestForApprovalCommand

func (CancelPurchaseRequestCommand) CqrsRequestType() cqrs.RequestType {
	return cancelPurchaseRequestCommandType
}

type CancelPurchaseRequestResult = dyn.OpResult[dyn.MutateResultData]

var markPurchaseRequestPriorityCommandType = cqrs.RequestType{
	Module: "purchase", Submodule: "purchaserequest", Action: "mark_priority",
}

type MarkPurchaseRequestPriorityCommand struct {
	Id       model.Id `json:"id" param:"id"`
	Etag     string   `json:"etag"`
	Priority string   `json:"priority"`
}

func (MarkPurchaseRequestPriorityCommand) CqrsRequestType() cqrs.RequestType {
	return markPurchaseRequestPriorityCommandType
}

type MarkPurchaseRequestPriorityResult = dyn.OpResult[dyn.MutateResultData]

var convertPurchaseRequestToRfqCommandType = cqrs.RequestType{
	Module: "purchase", Submodule: "purchaserequest", Action: "convert_rfq",
}

type ConvertPurchaseRequestToRfqCommand submitPurchaseRequestForApprovalCommand

func (ConvertPurchaseRequestToRfqCommand) CqrsRequestType() cqrs.RequestType {
	return convertPurchaseRequestToRfqCommandType
}

type ConvertPurchaseRequestToRfqResult = dyn.OpResult[dyn.MutateResultData]

var convertPurchaseRequestToPoCommandType = cqrs.RequestType{
	Module: "purchase", Submodule: "purchaserequest", Action: "convert_po",
}

type ConvertPurchaseRequestToPoCommand submitPurchaseRequestForApprovalCommand

func (ConvertPurchaseRequestToPoCommand) CqrsRequestType() cqrs.RequestType {
	return convertPurchaseRequestToPoCommandType
}

type ConvertPurchaseRequestToPoResult = dyn.OpResult[dyn.MutateResultData]

var consolidatePurchaseRequestsCommandType = cqrs.RequestType{
	Module: "purchase", Submodule: "purchaserequest", Action: "consolidate",
}

type ConsolidatePurchaseRequestsCommand struct {
	SourcePurchaseRequestIds []model.Id `json:"source_purchase_request_ids"`
	TargetPurchaseOrderId    *model.Id  `json:"target_purchase_order_id"`
	Etag                     string     `json:"etag"`
}

func (ConsolidatePurchaseRequestsCommand) CqrsRequestType() cqrs.RequestType {
	return consolidatePurchaseRequestsCommandType
}

type ConsolidatePurchaseRequestsResult = dyn.OpResult[dyn.MutateResultData]

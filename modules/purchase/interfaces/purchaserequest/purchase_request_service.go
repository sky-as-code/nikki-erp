package purchaserequest

import corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

type PurchaseRequestService interface {
	CreatePurchaseRequest(ctx corectx.Context, cmd CreatePurchaseRequestCommand) (*CreatePurchaseRequestResult, error)
	DeletePurchaseRequest(ctx corectx.Context, cmd DeletePurchaseRequestCommand) (*DeletePurchaseRequestResult, error)
	PurchaseRequestExists(ctx corectx.Context, query PurchaseRequestExistsQuery) (*PurchaseRequestExistsResult, error)
	GetPurchaseRequest(ctx corectx.Context, query GetPurchaseRequestQuery) (*GetPurchaseRequestResult, error)
	SearchPurchaseRequests(ctx corectx.Context, query SearchPurchaseRequestsQuery) (*SearchPurchaseRequestsResult, error)
	SetPurchaseRequestIsArchived(
		ctx corectx.Context, cmd SetPurchaseRequestIsArchivedCommand,
	) (*SetPurchaseRequestIsArchivedResult, error)
	UpdatePurchaseRequest(ctx corectx.Context, cmd UpdatePurchaseRequestCommand) (*UpdatePurchaseRequestResult, error)
	SubmitPurchaseRequestForApproval(
		ctx corectx.Context, cmd SubmitPurchaseRequestForApprovalCommand,
	) (*SubmitPurchaseRequestForApprovalResult, error)
	ApprovePurchaseRequest(ctx corectx.Context, cmd ApprovePurchaseRequestCommand) (*ApprovePurchaseRequestResult, error)
	RejectPurchaseRequest(ctx corectx.Context, cmd RejectPurchaseRequestCommand) (*RejectPurchaseRequestResult, error)
	CancelPurchaseRequest(ctx corectx.Context, cmd CancelPurchaseRequestCommand) (*CancelPurchaseRequestResult, error)
	MarkPurchaseRequestPriority(
		ctx corectx.Context, cmd MarkPurchaseRequestPriorityCommand,
	) (*MarkPurchaseRequestPriorityResult, error)
	ConvertPurchaseRequestToRfq(
		ctx corectx.Context, cmd ConvertPurchaseRequestToRfqCommand,
	) (*ConvertPurchaseRequestToRfqResult, error)
	ConvertPurchaseRequestToPo(
		ctx corectx.Context, cmd ConvertPurchaseRequestToPoCommand,
	) (*ConvertPurchaseRequestToPoResult, error)
	ConsolidatePurchaseRequests(
		ctx corectx.Context, cmd ConsolidatePurchaseRequestsCommand,
	) (*ConsolidatePurchaseRequestsResult, error)
}

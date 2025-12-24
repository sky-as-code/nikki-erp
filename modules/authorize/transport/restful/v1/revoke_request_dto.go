package v1

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/revoke_request"
)

type RevokeRequestDto struct {
	Id   model.Id   `json:"id"`
	Etag model.Etag `json:"etag"`

	AttachmentURL *string   `json:"attachmentUrl,omitempty"`
	Comment       *string   `json:"comment,omitempty"`
	TargetType    string    `json:"targetType,omitempty"`
	ReceiverType  string    `json:"receiverType,omitempty"`
	ResponseId    *model.Id `json:"responseId,omitempty"`
	Status        string    `json:"status,omitempty"`

	Requestor *UserSummaryDto `json:"requestor,omitempty" model:"-"`
	Receiver  *UserSummaryDto `json:"receiver,omitempty" model:"-"`
	Target    *TargetSummaryDto `json:"target,omitempty" model:"-"`
}

type RevokeRequestSummaryDto struct {
	Id   model.Id `json:"id"`
	Name string   `json:"name"`
}

func (this *RevokeRequestDto) FromRevokeRequest(revokeRequest domain.RevokeRequest) {
	model.MustCopy(revokeRequest.ModelBase, this)
	model.MustCopy(revokeRequest.AuditableBase, this)
	model.MustCopy(revokeRequest, this)

	this.Requestor = &UserSummaryDto{}
	this.Requestor.FromUserSummary(*revokeRequest.RequestorId, revokeRequest.RequestorName)

	this.Receiver = &UserSummaryDto{}
	this.Receiver.FromUserSummary(*revokeRequest.ReceiverId, revokeRequest.ReceiverName)

	if revokeRequest.Role != nil {
		this.Target = &TargetSummaryDto{}
		this.Target.FromTargetSummary(*revokeRequest.TargetRef, revokeRequest.Role.Name)
	} else if revokeRequest.RoleSuite != nil {
		this.Target = &TargetSummaryDto{}
		this.Target.FromTargetSummary(*revokeRequest.TargetRef, revokeRequest.RoleSuite.Name)
	}
}

type CreateRevokeRequestRequest = it.CreateRevokeRequestCommand
type CreateRevokeRequestResponse = httpserver.RestCreateResponse

type CreateBulkRevokeRequestsRequest = it.CreateBulkRevokeRequestsCommand

type CreateBulkRevokeRequestsResponse struct {
	Items []httpserver.RestCreateResponse `json:"items"`
}

type GetRevokeRequestByIdRequest = it.GetRevokeRequestByIdQuery
type GetRevokeRequestByIdResponse = RevokeRequestDto

type SearchRevokeRequestsRequest = it.SearchRevokeRequestsQuery
type SearchRevokeRequestsResponse httpserver.RestSearchResponse[RevokeRequestDto]

func (this *SearchRevokeRequestsResponse) FromResult(result *it.SearchRevokeRequestsResultData) {
	this.Total = result.Total
	this.Page = result.Page
	this.Size = result.Size
	this.Items = array.Map(result.Items, func(revokeRequest domain.RevokeRequest) RevokeRequestDto {
		item := RevokeRequestDto{}
		item.FromRevokeRequest(revokeRequest)
		return item
	})
}

type DeleteRevokeRequestRequest = it.DeleteRevokeRequestCommand
type DeleteRevokeRequestResponse = httpserver.RestDeleteResponse

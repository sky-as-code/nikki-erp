package v1

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/grant_request"
)

type GrantRequestDto struct {
	Id   model.Id   `json:"id"`
	Etag model.Etag `json:"etag"`

	AttachmentUrl *string   `json:"attachmentUrl,omitempty"`
	Comment       *string   `json:"comment,omitempty"`
	TargetType    string    `json:"targetType,omitempty"`
	TargetRef     *model.Id `json:"targetRef,omitempty"`
	ResponseId    *model.Id `json:"responseId,omitempty"`
	Status        string    `json:"status,omitempty"`

	GrantResponses []GrantResponseDto `json:"grantResponses,omitempty"`
	Receiver       *UserSummaryDto    `json:"receiver,omitempty"`
	Requestor      *UserSummaryDto    `json:"requestor,omitempty"`
	Target         *TargetSummaryDto  `json:"target,omitempty"`
}

type GrantRequestSummaryDto struct {
	Id   model.Id `json:"id"`
	Name string   `json:"name"`
}

func (this *GrantRequestDto) FromGrantRequest(grantRequest domain.GrantRequest) {
	model.MustCopy(grantRequest.ModelBase, this)
	model.MustCopy(grantRequest.AuditableBase, this)
	model.MustCopy(grantRequest, this)

	this.Requestor = &UserSummaryDto{}
	this.Requestor.FromUserSummary(*grantRequest.RequestorId, grantRequest.RequestorName)

	this.Receiver = &UserSummaryDto{}
	this.Receiver.FromUserSummary(*grantRequest.ReceiverId, grantRequest.ReceiverName)

	this.Target = &TargetSummaryDto{}
	if grantRequest.Role != nil {
		this.Target.FromTargetSummary(*grantRequest.TargetRef, grantRequest.Role.Name)
	} else if grantRequest.RoleSuite != nil {
		this.Target.FromTargetSummary(*grantRequest.TargetRef, grantRequest.RoleSuite.Name)
	}

	if grantRequest.GrantResponses != nil {
		this.GrantResponses = array.Map(grantRequest.GrantResponses, func(grantResponse domain.GrantResponse) GrantResponseDto {
			grantResponseDto := GrantResponseDto{}
			grantResponseDto.FromGrantResponse(grantResponse)
			return grantResponseDto
		})
	}
}

type GrantResponseDto struct {
	Id            model.Id `json:"id"`
	ResponderName *string  `json:"responderName,omitempty"`
	IsApproved    *bool    `json:"isApproved,omitempty"`
}

func (this *GrantResponseDto) FromGrantResponse(grantResponse domain.GrantResponse) {
	model.MustCopy(grantResponse.ModelBase, this)
	model.MustCopy(grantResponse.AuditableBase, this)
	model.MustCopy(grantResponse, this)
}

type CreateGrantRequestRequest = it.CreateGrantRequestCommand
type CreateGrantRequestResponse = httpserver.RestCreateResponse

type CancelGrantRequestRequest = it.CancelGrantRequestCommand
type CancelGrantRequestResponse = httpserver.RestUpdateResponse

type DeleteGrantRequestRequest = it.DeleteGrantRequestCommand
type DeleteGrantRequestResponse = httpserver.RestDeleteResponse

type GetGrantRequestByIdRequest = it.GetGrantRequestByIdQuery
type GetGrantRequestByIdResponse = GrantRequestDto

type RespondToGrantRequestRequest = it.RespondToGrantRequestCommand
type RespondToGrantRequestResponse = httpserver.RestUpdateResponse

// type SearchGrantRequestsRequest = it.SearchGrantRequestsCommand
// type SearchGrantRequestsResponse httpserver.RestSearchResponse[GrantRequestDto]

// func (this *SearchGrantRequestsResponse) FromResult(result *it.SearchGrantRequestsResultData) {
// 	this.Total = result.Total
// 	this.Page = result.Page
// 	this.Size = result.Size
// 	this.Items = array.Map(result.Items, func(grantRequest domain.GrantRequest) GrantRequestDto {
// 		item := GrantRequestDto{}
// 		item.FromGrantRequest(grantRequest)
// 		return item
// 	})
// }

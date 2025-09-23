package v1

import (
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
	ApprovalId    *model.Id `json:"approvalId,omitempty"`
	RequestorId   *model.Id `json:"requestorId,omitempty"`
	ReceiverId    *model.Id `json:"receiverId,omitempty"`
	TargetType    string    `json:"targetType,omitempty"`
	TargetRef     *model.Id `json:"targetRef,omitempty"`
	ResponseId    *model.Id `json:"responseId,omitempty"`
	Status        string    `json:"status,omitempty"`

	Role      *RoleDto      `json:"role,omitempty"`
	RoleSuite *RoleSuiteDto `json:"roleSuite,omitempty"`
}

type GrantRequestSummaryDto struct {
	Id   model.Id `json:"id"`
	Name string   `json:"name"`
}

func (this *GrantRequestDto) FromGrantRequest(grantRequest domain.GrantRequest) {
	model.MustCopy(grantRequest.ModelBase, this)
	model.MustCopy(grantRequest.AuditableBase, this)
	model.MustCopy(grantRequest, this)

	if grantRequest.Role != nil {
		this.Role = &RoleDto{}
		this.Role.FromRole(*grantRequest.Role)
	}

	if grantRequest.RoleSuite != nil {
		this.RoleSuite = &RoleSuiteDto{}
		this.RoleSuite.FromRoleSuite(*grantRequest.RoleSuite)
	}
}

// func (this *GrantRequestSummaryDto) FromGrantRequest(grantRequest domain.GrantRequest) {
// 	this.Id = *grantRequest.Id
// 	this.Name = *grantRequest.TargetType.String()
// }

type CreateGrantRequestRequest = it.CreateGrantRequestCommand
type CreateGrantRequestResponse = httpserver.RestCreateResponse

type CancelGrantRequestRequest = it.CancelGrantRequestCommand
type CancelGrantRequestResponse = httpserver.RestUpdateResponse

type DeleteGrantRequestRequest = it.DeleteGrantRequestCommand
type DeleteGrantRequestResponse = httpserver.RestDeleteResponse

type RespondToGrantRequestRequest = it.RespondToGrantRequestCommand
type RespondToGrantRequestResponse = httpserver.RestUpdateResponse

// type GetGrantRequestByIdRequest = it.GetGrantRequestByIdQuery
// type GetGrantRequestByIdResponse = GrantRequestDto

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

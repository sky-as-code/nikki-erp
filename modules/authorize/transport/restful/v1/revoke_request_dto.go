package v1

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/revoke_request"
)

type RevokeRequestDto struct {
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
}

type RevokeRequestSummaryDto struct {
	Id   model.Id `json:"id"`
	Name string   `json:"name"`
}

func (this *RevokeRequestDto) FromRevokeRequest(revokeRequest domain.RevokeRequest) {
	model.MustCopy(revokeRequest.ModelBase, this)
	model.MustCopy(revokeRequest.AuditableBase, this)
	model.MustCopy(revokeRequest, this)
}

type CreateRevokeRequestRequest = it.CreateRevokeRequestCommand
type CreateRevokeRequestResponse = httpserver.RestCreateResponse


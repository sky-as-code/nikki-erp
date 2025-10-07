package revoke_request

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

func init() {
	// Assert interface implementation
	var req cqrs.Request
	util.Unused(req)
}

// START: CreateRevokeRequestCommand
var createRevokeRequestCommandType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "revoke_request",
	Action:    "create",
}

type CreateRevokeRequestCommand struct {
	AttachmentUrl *string                       `json:"attachmentUrl"`
	Comment       *string                       `json:"comment"`
	RequestorId   model.Id                      `json:"requestorId"`
	ReceiverType  domain.ReceiverType           `json:"receiverType"`
	ReceiverId    model.Id                      `json:"receiverId"`
	TargetType    domain.GrantRequestTargetType `json:"targetType"`
	TargetRef     model.Id                      `json:"targetRef"`
}

func (CreateRevokeRequestCommand) CqrsRequestType() cqrs.RequestType {
	return createRevokeRequestCommandType
}

type CreateRevokeRequestResult = crud.OpResult[*domain.RevokeRequest]

// END: CreateRevokeRequestCommand

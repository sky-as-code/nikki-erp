package revoke_request

import (
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"

	itGrantRequest "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/grant_request"
)

func init() {
	// Assert interface implementation
	var req cqrs.Request
	req = (*TargetIsDeletedCommand)(nil)
	req = (*GetRevokeRequestByIdQuery)(nil)
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

// START: TargetIsDeleted
var targetIsDeletedCommandType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "revoke_request",
	Action:    "targetIsDeleted",
}

type TargetIsDeletedCommand struct {
	TargetType domain.GrantRequestTargetType `json:"targetType"`
	TargetRef  model.Id                      `json:"targetRef"`
	TargetName string                        `json:"targetName"`
}

func (TargetIsDeletedCommand) CqrsRequestType() cqrs.RequestType {
	return targetIsDeletedCommandType
}

func (this TargetIsDeletedCommand) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.TargetRef, true),
		validator.Field(&this.TargetName, validator.NotEmpty),
		itGrantRequest.GrantRequestTargetTypeValidateRule(&this.TargetType),
	}

	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type TargetIsDeletedResult = crud.OpResult[bool]

// END: TargetIsDeleted

// START: GetRevokeRequestByIdQuery
var getRevokeRequestByIdQueryType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "revoke_request",
	Action:    "getById",
}

type GetRevokeRequestByIdQuery struct {
	Id model.Id `param:"id" json:"id"`
}

func (GetRevokeRequestByIdQuery) CqrsRequestType() cqrs.RequestType {
	return getRevokeRequestByIdQueryType
}

func (this GetRevokeRequestByIdQuery) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}

	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type GetRevokeRequestByIdResult = crud.OpResult[*domain.RevokeRequest]

// END: GetRevokeRequestByIdQuery

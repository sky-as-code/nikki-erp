package grant_request

import (
	"time"

	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

func init() {
	// Assert interface implementation
	var req cqrs.Request
	req = (*CreateGrantRequestCommand)(nil)
	req = (*GetGrantRequestQuery)(nil)
	req = (*RespondToGrantRequestCommand)(nil)
	util.Unused(req)
}

// START: CreateGrantRequestCommand
var createGrantRequestCommandType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "grant_request",
	Action:    "create",
}

type CreateGrantRequestCommand struct {
	AttachmentUrl *string                       `json:"attachmentUrl"`
	Comment       *string                       `json:"comment"`
	RequestorId   model.Id                      `json:"requestorId"`
	ReceiverType  domain.ReceiverType           `json:"receiverType"`
	ReceiverId    model.Id                      `json:"receiverId"`
	TargetType    domain.GrantRequestTargetType `json:"targetType"`
	TargetRef     model.Id                      `json:"targetRef"`
}

func (CreateGrantRequestCommand) CqrsRequestType() cqrs.RequestType {
	return createGrantRequestCommandType
}

func (this CreateGrantRequestCommand) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		validator.Field(&this.AttachmentUrl,
			validator.When(this.AttachmentUrl != nil,
				validator.NotEmpty,
				validator.Length(1, model.MODEL_RULE_URL_LENGTH),
			),
		),
		validator.Field(&this.Comment,
			validator.When(this.Comment != nil,
				validator.NotEmpty,
				validator.Length(1, model.MODEL_RULE_DESC_LENGTH),
			),
		),
		GrantRequestTargetTypeValidateRule(&this.TargetType),
		ReceiverTypeValidateRule(&this.ReceiverType),
		model.IdValidateRule(&this.RequestorId, true),
		model.IdValidateRule(&this.ReceiverId, true),
		model.IdValidateRule(&this.TargetRef, true),
	}

	return validator.ApiBased.ValidateStruct(&this, rules...)
}

func GrantRequestTargetTypeValidateRule(field *domain.GrantRequestTargetType) *validator.FieldRules {
	return validator.Field(field,
		validator.NotEmpty,
		validator.OneOf(
			domain.GrantRequestTargetTypeRole,
			domain.GrantRequestTargetTypeSuite,
		),
	)
}

func ReceiverTypeValidateRule(field *domain.ReceiverType) *validator.FieldRules {
	return validator.Field(field,
		validator.NotEmpty,
		validator.OneOf(
			domain.ReceiverTypeUser,
			domain.ReceiverTypeGroup,
		),
	)
}

type CreateGrantRequestResult = crud.OpResult[*domain.GrantRequest]

// END: CreateGrantRequestCommand

// START: Exist

// START: CancelGrantRequestCommand
var cancelGrantRequestCommandType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "grant_request",
	Action:    "cancel",
}

type CancelGrantRequestCommand struct {
	Id *model.Id `json:"id"`
}

func (CancelGrantRequestCommand) CqrsRequestType() cqrs.RequestType {
	return cancelGrantRequestCommandType
}

type CancelGrantRequestResult = crud.OpResult[*time.Time]

// END: CancelGrantRequestCommand

// START: RespondToGrantRequestCommand
var respondToGrantRequestCommandType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "grant_request",
	Action:    "respond",
}

type RespondToGrantRequestCommand struct {
	Id          model.Id                    `param:"id" json:"id"`
	Decision    domain.GrantRequestDecision `json:"decision"`
	ResponderId model.Id                    `json:"responderId"`
}

func (RespondToGrantRequestCommand) CqrsRequestType() cqrs.RequestType {
	return respondToGrantRequestCommandType
}

func (this RespondToGrantRequestCommand) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.Id, true),
		validator.Field(&this.Decision,
			validator.NotEmpty,
			validator.OneOf(
				domain.GrantRequestDecisionApprove,
				domain.GrantRequestDecisionDeny,
			),
		),
		model.IdValidateRule(&this.ResponderId, true),
	}

	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type RespondToGrantRequestResult = crud.OpResult[*domain.GrantRequest]

// END: RespondToGrantRequestCommand

// START: GetGrantRequestQuery
var getGrantRequestQueryType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "grant_request",
	Action:    "get",
}

type GetGrantRequestQuery struct {
	Id model.Id `param:"id" json:"id"`
}

func (GetGrantRequestQuery) CqrsRequestType() cqrs.RequestType {
	return getGrantRequestQueryType
}

type GetGrantRequestResult = crud.OpResult[*domain.GrantRequest]

// END: GetGrantRequestQuery

// START: ListGrantRequestsQuery
type ListGrantRequestsQuery struct {
	Page       *int                    `json:"page,omitempty"`
	Size       *int                    `json:"size,omitempty"`
	Graph      *map[string]interface{} `json:"graph,omitempty"`
	Status     *string                 `json:"status,omitempty"`
	ApprovalId *model.Id               `json:"approvalId,omitempty"`
}

type ListGrantRequestsResult = crud.OpResult[[]*domain.GrantRequest]

// END: ListGrantRequestsQuery

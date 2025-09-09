package domain

import (
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/validator"
	entGrantRequest "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/grantrequest"
)

type GrantRequest struct {
	model.ModelBase
	model.AuditableBase

	AttachmentUrl   *string                 `json:"attachmentUrl,omitempty"`
	Comment         *string                 `json:"comment,omitempty"`
	ApprovalId      *model.Id               `json:"approvalId,omitempty"`
	RequestorId     *model.Id               `json:"requestorId,omitempty"`
	ReceiverId      *model.Id               `json:"receiverId,omitempty"`
	ReceiverType    *ReceiverType           `json:"receiverType,omitempty"`
	TargetType      *GrantRequestTargetType `json:"targetType,omitempty"`
	TargetRef       *model.Id               `json:"targetRef,omitempty"`
	ResponseId      *model.Id               `json:"responseId,omitempty"` // Only set after response
	Status          *GrantRequestStatus     `json:"status,omitempty"`
	TargetRoleName  *string                 `json:"targetRoleName,omitempty"`
	TargetSuiteName *string                 `json:"targetSuiteName,omitempty"`

	Role           *Role           `json:"role,omitempty" model:"-"` // TODO: Handle copy
	RoleSuite      *RoleSuite      `json:"roleSuite,omitempty" model:"-"`
	GrantResponses []GrantResponse `json:"grantResponses,omitempty" model:"-"`
}

func (this *GrantRequest) Validate(forEdit bool) fault.ValidationErrors {
	rules := []*validator.FieldRules{
		validator.Field(&this.TargetType,
			validator.NotNilWhen(!forEdit),
			validator.When(this.TargetType != nil,
				validator.NotEmpty,
			),
		),
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
		GrantRequestTargetTypeValidateRule(&this.TargetType, !forEdit),
		GrantRequestStatusValidateRule(&this.Status, !forEdit),
		ReceiverTypeValidateRule(&this.ReceiverType, !forEdit),
		model.IdPtrValidateRule(&this.RequestorId, !forEdit),
		model.IdPtrValidateRule(&this.ReceiverId, !forEdit),
		model.IdPtrValidateRule(&this.TargetRef, !forEdit),
		model.IdPtrValidateRule(&this.ApprovalId, !forEdit),
		model.IdPtrValidateRule(&this.ResponseId, !forEdit),
	}
	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)

	return validator.ApiBased.ValidateStruct(this, rules...)
}

type GrantRequestTargetType entGrantRequest.TargetType

const (
	GrantRequestTargetTypeRole  = GrantRequestTargetType(entGrantRequest.TargetTypeRole)
	GrantRequestTargetTypeSuite = GrantRequestTargetType(entGrantRequest.TargetTypeSuite)
)

func (this GrantRequestTargetType) String() string {
	return string(this)
}

func WrapGrantTargetType(s string) *GrantRequestTargetType {
	st := GrantRequestTargetType(s)
	return &st
}

func WrapGrantRequestTargetTypeEnt(s entGrantRequest.TargetType) *GrantRequestTargetType {
	st := GrantRequestTargetType(s)
	return &st
}

func GrantRequestTargetTypeValidateRule(field **GrantRequestTargetType, isRequired bool) *validator.FieldRules {
	return validator.Field(field,
		validator.NotNilWhen(isRequired),
		validator.When(*field != nil,
			validator.NotEmpty,
			validator.OneOf(GrantRequestTargetTypeRole, GrantRequestTargetTypeSuite),
		),
	)
}

type GrantRequestStatus entGrantRequest.Status

const (
	PendingGrantRequestStatus   = GrantRequestStatus(entGrantRequest.StatusPending)
	ApprovedGrantRequestStatus  = GrantRequestStatus(entGrantRequest.StatusApproved)
	RejectedGrantRequestStatus  = GrantRequestStatus(entGrantRequest.StatusRejected)
	CancelledGrantRequestStatus = GrantRequestStatus(entGrantRequest.StatusCancelled)
)

func (this GrantRequestStatus) String() string {
	return string(this)
}

func WrapGrantRequestStatus(s string) *GrantRequestStatus {
	st := GrantRequestStatus(s)
	return &st
}

func WrapGrantRequestStatusEnt(s entGrantRequest.Status) *GrantRequestStatus {
	st := GrantRequestStatus(s)
	return &st
}

func GrantRequestStatusValidateRule(field **GrantRequestStatus, isRequired bool) *validator.FieldRules {
	return validator.Field(field,
		validator.NotNilWhen(isRequired),
		validator.When(*field != nil,
			validator.NotEmpty,
			validator.OneOf(
				PendingGrantRequestStatus,
				ApprovedGrantRequestStatus,
				RejectedGrantRequestStatus,
				CancelledGrantRequestStatus,
			),
		),
	)
}

type GrantRequestDecision string

const (
	GrantRequestDecisionApprove = GrantRequestDecision("approve")
	GrantRequestDecisionDeny    = GrantRequestDecision("deny")
)

type ReceiverType entGrantRequest.ReceiverType

const (
	ReceiverTypeUser  = ReceiverType(entGrantRequest.ReceiverTypeUser)
	ReceiverTypeGroup = ReceiverType(entGrantRequest.ReceiverTypeGroup)
)

func (this ReceiverType) String() string {
	return string(this)
}

func WrapReceiverType(s string) *ReceiverType {
	rt := ReceiverType(s)
	return &rt
}

func WrapReceiverTypeEnt(s entGrantRequest.ReceiverType) *ReceiverType {
	rt := ReceiverType(s)
	return &rt
}

func ReceiverTypeValidateRule(field **ReceiverType, isRequired bool) *validator.FieldRules {
	return validator.Field(field,
		validator.NotNilWhen(isRequired),
		validator.When(*field != nil,
			validator.NotEmpty,
			validator.OneOf(ReceiverTypeUser, ReceiverTypeGroup),
		),
	)
}

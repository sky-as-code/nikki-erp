package domain

import (
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/validator"
	entRevokeRequest "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/revokerequest"
)

type RevokeRequest struct {
	model.ModelBase
	model.AuditableBase

	AttachmentURL   *string                  `json:"attachmentUrl,omitempty"`
	Comment         *string                  `json:"comment,omitempty"`
	RequestorId     *model.Id                `json:"requestorId,omitempty"`
	RequestorName   *string                  `json:"requestorName,omitempty"`
	ReceiverType    *ReceiverType            `json:"receiverType,omitempty"`
	ReceiverId      *model.Id                `json:"receiverId,omitempty"`
	ReceiverName    *string                  `json:"receiverName,omitempty"`
	TargetType      *RevokeRequestTargetType `json:"targetType,omitempty"`
	TargetRef       *model.Id                `json:"targetRef,omitempty"`
	TargetRoleName  *string                  `json:"targetRoleName,omitempty"`  // Only set after role is deleted
	TargetSuiteName *string                  `json:"targetSuiteName,omitempty"` // Only set after suite is deleted

	Role      *Role      `json:"role,omitempty" model:"-"` // TODO: Handle copy
	RoleSuite *RoleSuite `json:"roleSuite,omitempty" model:"-"`
}

func (this *RevokeRequest) Validate(forEdit bool) fault.ValidationErrors {
	rules := []*validator.FieldRules{
		validator.Field(&this.AttachmentURL,
			validator.NotNilWhen(!forEdit),
			validator.When(this.AttachmentURL != nil,
				validator.NotEmpty,
				validator.Length(1, model.MODEL_RULE_URL_LENGTH),
			),
		),
		validator.Field(&this.Comment,
			validator.NotNilWhen(!forEdit),
			validator.When(this.Comment != nil,
				validator.NotEmpty,
				validator.Length(1, model.MODEL_RULE_DESC_LENGTH),
			),
		),
		RevokeRequestTargetTypeValidateRule(&this.TargetType, !forEdit),
		ReceiverTypeValidateRule(&this.ReceiverType, !forEdit),
		model.IdPtrValidateRule(&this.RequestorId, !forEdit),
		model.IdPtrValidateRule(&this.ReceiverId, !forEdit),
		model.IdPtrValidateRule(&this.TargetRef, !forEdit),
	}

	return validator.ApiBased.ValidateStruct(this, rules...)
}

type RevokeRequestTargetType entRevokeRequest.TargetType

const (
	RevokeRequestTargetTypeNikkiRole  = RevokeRequestTargetType(entRevokeRequest.TargetTypeRole)
	RevokeRequestTargetTypeNikkiSuite = RevokeRequestTargetType(entRevokeRequest.TargetTypeSuite)
)

func (this RevokeRequestTargetType) String() string {
	return string(this)
}

func WrapRevokeTargetType(s string) *RevokeRequestTargetType {
	st := RevokeRequestTargetType(s)
	return &st
}

func WrapRevokeRequestTargetTypeEnt(s entRevokeRequest.TargetType) *RevokeRequestTargetType {
	st := RevokeRequestTargetType(s)
	return &st
}

func RevokeRequestTargetTypeValidateRule(field **RevokeRequestTargetType, isRequired bool) *validator.FieldRules {
	return validator.Field(field,
		validator.NotNilWhen(isRequired),
		validator.When(*field != nil,
			validator.NotEmpty,
			validator.OneOf(
				RevokeRequestTargetTypeNikkiRole,
				RevokeRequestTargetTypeNikkiSuite,
			),
		),
	)
}

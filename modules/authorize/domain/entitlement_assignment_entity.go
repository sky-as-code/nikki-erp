package domain

import (
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/validator"

	entEntitlementAssignment "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/entitlementassignment"
)

type EntitlementAssignment struct {
	model.ModelBase

	SubjectType   *EntitlementAssignmentSubjectType `json:"subjectType,omitempty"`
	SubjectRef    *string                           `json:"subjectRef,omitempty"`
	ActionName    *string                           `json:"actionName,omitempty"`
	ResourceName  *string                           `json:"resourceName,omitempty"`
	ResolvedExpr  *string                           `json:"resolvedExpr,omitempty"`
	EntitlementId *model.Id                         `json:"entitlementId,omitempty"`

	Entitlement *Entitlement `json:"entitlement,omitempty" model:"-"` // TODO: Handle copy
	Role        *Role        `json:"role,omitempty" model:"-"`
}

func (this *EntitlementAssignment) Validate(forEdit bool) fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdPtrValidateRule(&this.EntitlementId, !forEdit),
		validator.Field(&this.ActionName,
			validator.NotNilWhen(!forEdit),
			validator.When(this.ActionName != nil,
				validator.NotEmpty,
			),
		),
		validator.Field(&this.ResourceName,
			validator.NotNilWhen(!forEdit),
			validator.When(this.ResourceName != nil,
				validator.NotEmpty,
				validator.Length(1, model.MODEL_RULE_SHORT_NAME_LENGTH),
			),
		),
		validator.Field(&this.ResolvedExpr,
			validator.When(this.ResolvedExpr != nil,
				validator.NotEmpty,
				validator.Length(1, model.MODEL_RULE_DESC_LENGTH),
			),
		),
		validator.Field(&this.SubjectType,
			validator.NotEmpty,
			validator.OneOf(
				EntitlementAssignmentSubjectTypeNikkiUser,
				EntitlementAssignmentSubjectTypeNikkiGroup,
				EntitlementAssignmentSubjectTypeNikkiRole,
				EntitlementAssignmentSubjectTypeCustom,
			),
		),
		model.IdPtrValidateRule(&this.SubjectRef, !forEdit),
	}
	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)

	return validator.ApiBased.ValidateStruct(this, rules...)
}

type EntitlementAssignmentSubjectType entEntitlementAssignment.SubjectType

const (
	EntitlementAssignmentSubjectTypeNikkiUser  = EntitlementAssignmentSubjectType(entEntitlementAssignment.SubjectTypeNikkiUser)
	EntitlementAssignmentSubjectTypeNikkiGroup = EntitlementAssignmentSubjectType(entEntitlementAssignment.SubjectTypeNikkiGroup)
	EntitlementAssignmentSubjectTypeNikkiRole  = EntitlementAssignmentSubjectType(entEntitlementAssignment.SubjectTypeNikkiRole)
	EntitlementAssignmentSubjectTypeCustom     = EntitlementAssignmentSubjectType(entEntitlementAssignment.SubjectTypeCustom)
)

func (this EntitlementAssignmentSubjectType) String() string {
	return string(this)
}

func WrapEntitlementAssignmentSubjectType(s string) *EntitlementAssignmentSubjectType {
	st := EntitlementAssignmentSubjectType(s)
	return &st
}

func WrapEntitlementAssignmentSubjectTypeEnt(s entEntitlementAssignment.SubjectType) *EntitlementAssignmentSubjectType {
	st := EntitlementAssignmentSubjectType(s)
	return &st
}

package domain

import (
	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	entEntitlement "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/entitlement"
)

type Entitlement struct {
	model.ModelBase
	model.AuditableBase

	ActionId    *model.Id               `json:"actionId,omitempty"`
	ActionExpr  *string                 `json:"actionExpr,omitempty"`
	Description *string                 `json:"description,omitempty"`
	Name        *string                 `json:"name,omitempty"`
	ResourceId  *model.Id               `json:"resourceId,omitempty"`
	SubjectType *EntitlementSubjectType `json:"subjectType,omitempty"`
	SubjectRef  *string                 `json:"subjectRef,omitempty"`
	ScopeRef    *string                 `json:"scopeRef,omitempty"`

	Action   *Action   `json:"action,omitempty"`
	Resource *Resource `json:"resource,omitempty"`
}

func (this *Entitlement) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdPtrValidateRule(&this.ActionId, true),
		val.Field(&this.ActionExpr,
			val.NotEmpty,
		),
		val.Field(&this.Name,
			val.NotNilWhen(!forEdit),
			val.When(this.Name != nil,
				val.NotEmpty,
				val.Length(1, model.MODEL_RULE_SHORT_NAME_LENGTH),
			),
		),
		val.Field(&this.Description,
			val.When(this.Description != nil,
				val.NotEmpty,
				val.Length(1, model.MODEL_RULE_DESC_LENGTH),
			),
		),
		model.IdPtrValidateRule(&this.ResourceId, true),
		EntitlementSubjectTypeValidateRule(&this.SubjectType),

		val.Field(&this.SubjectRef,
			val.NotEmpty,
			val.Length(1, model.MODEL_RULE_NON_NIKKI_ID_LENGTH),
		),
		val.Field(&this.ScopeRef,
			val.NotEmpty,
			val.Length(1, model.MODEL_RULE_LONG_NAME_LENGTH),
		),
	}
	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)

	return val.ApiBased.ValidateStruct(this, rules...)
}

type EntitlementSubjectType entEntitlement.SubjectType

const (
	EntitlementSubjectTypeNikkiUser  = EntitlementSubjectType(entEntitlement.SubjectTypeNikkiUser)
	EntitlementSubjectTypeNikkiGroup = EntitlementSubjectType(entEntitlement.SubjectTypeNikkiGroup)
	EntitlementSubjectTypeNikkiRole  = EntitlementSubjectType(entEntitlement.SubjectTypeNikkiRole)
	EntitlementSubjectTypeCustom     = EntitlementSubjectType(entEntitlement.SubjectTypeCustom)
)

func (this EntitlementSubjectType) Validate() error {
	switch this {
	case EntitlementSubjectTypeNikkiUser, EntitlementSubjectTypeNikkiGroup, EntitlementSubjectTypeNikkiRole, EntitlementSubjectTypeCustom:
		return nil
	default:
		return errors.Errorf("invalid subject type value: %s", this)
	}
}

func (this EntitlementSubjectType) String() string {
	return string(this)
}

func WrapEntitlementSubjectType(s string) *EntitlementSubjectType {
	st := EntitlementSubjectType(s)
	return &st
}

func WrapEntitlementSubjectTypeEnt(s entEntitlement.SubjectType) *EntitlementSubjectType {
	st := EntitlementSubjectType(s)
	return &st
}

func EntitlementSubjectTypeValidateRule(field **EntitlementSubjectType) *val.FieldRules {
	return val.Field(field,
		val.NotEmpty,
		val.OneOf(EntitlementSubjectTypeNikkiUser, EntitlementSubjectTypeCustom),
	)
}

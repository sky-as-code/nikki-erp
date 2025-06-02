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
	ResourceId  *model.Id               `json:"resourceId,omitempty"`
	SubjectType *EntitlementSubjectType `json:"subjectType,omitempty"`
	SubjectRef  *string                 `json:"subjectRef,omitempty"`
	ScopeRef    *string                 `json:"scopeRef,omitempty"`

	Action   *Action   `json:"action,omitempty"`
	Resource *Resource `json:"resource,omitempty"`
}

func (this *Entitlement) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.ActionId, true),
		val.Field(&this.ActionExpr,
			val.Required,
		),
		model.IdValidateRule(&this.ResourceId, true),
		EntitlementSubjectTypeValidateRule(&this.SubjectType),

		val.Field(&this.SubjectRef,
			val.Required,
			val.Length(1, model.MODEL_RULE_NON_NIKKI_ID_LENGTH),
		),
		val.Field(&this.ScopeRef,
			val.Required,
			val.Length(1, model.MODEL_RULE_LONG_NAME_LENGTH),
		),
	}
	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)

	return val.ApiBased.ValidateStruct(this, rules...)
}

type EntitlementSubjectType entEntitlement.SubjectType

const (
	EntitlementSubjectTypeNikkiUser = EntitlementSubjectType(entEntitlement.SubjectTypeNikkiUser)
	EntitlementSubjectTypeCustom    = EntitlementSubjectType(entEntitlement.SubjectTypeCustom)
)

func (this EntitlementSubjectType) Validate() error {
	switch this {
	case EntitlementSubjectTypeNikkiUser, EntitlementSubjectTypeCustom:
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

func EntitlementSubjectTypeValidateRule(field any) *val.FieldRules {
	return val.Field(field,
		val.Required,
		val.OneOf(EntitlementSubjectTypeNikkiUser, EntitlementSubjectTypeCustom),
	)
}

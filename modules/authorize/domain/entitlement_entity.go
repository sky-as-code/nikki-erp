package domain

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	entEntitlement "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/entitlement"
	"go.bryk.io/pkg/errors"
)

type Entitlement struct {
	model.ModelBase

	SubjectType *EntitlementSubjectType `json:"subjectType,omitempty"`
	SubjectRef  *string                 `json:"subjectRef,omitempty"`
	ScopeRef    *string                 `json:"scopeRef,omitempty"`
	ActionId    *model.Id               `json:"actionId,omitempty"`
	ResourceId  *model.Id               `json:"resourceId,omitempty"`
}

func (this *Entitlement) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		EntitlementSubjectTypeValidateRule(&this.SubjectType),
		val.Field(&this.SubjectRef,
			val.Required,
			val.Length(1, 100),
		),
		val.Field(&this.ScopeRef,
			val.Required,
			val.Length(1, 100),
		),
		model.IdValidateRule(&this.ActionId, true),
		model.IdValidateRule(&this.ResourceId, true),
	}

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

package domain

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	// entEntitlement "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/entitlement"
)

type Entitlement struct {
	model.ModelBase
	model.AuditableBase

	ActionId    *model.Id               `json:"actionId,omitempty"`
	ActionExpr  *string                 `json:"actionExpr,omitempty"`
	Description *string                 `json:"description,omitempty"`
	Name        *string                 `json:"name,omitempty"`
	ResourceId  *model.Id               `json:"resourceId,omitempty"`
	// SubjectType *EntitlementSubjectType `json:"subjectType,omitempty"`
	// SubjectRef  *string                 `json:"subjectRef,omitempty"`
	ScopeRef    *string                 `json:"scopeRef,omitempty"`
	CreatedBy   *string                 `json:"createdBy,omitempty"`

	Action   *Action   `json:"action,omitempty"`
	Resource *Resource `json:"resource,omitempty"`
}

func (this *Entitlement) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdPtrValidateRule(&this.ActionId, false),
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
		// EntitlementSubjectTypeValidateRule(&this.SubjectType, !forEdit),
		// EntitlementSubjectRefValidateRule(&this.SubjectType, &this.SubjectRef, !forEdit),
		EntitlementScopeRefValidateRule(&this.ScopeRef),
		model.IdPtrValidateRule(&this.ResourceId, false),
		model.IdPtrValidateRule(&this.CreatedBy, !forEdit),
	}
	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)

	return val.ApiBased.ValidateStruct(this, rules...)
}

// type EntitlementSubjectType entEntitlement.SubjectType

// const (
// 	EntitlementSubjectTypeNikkiUser  = EntitlementSubjectType(entEntitlement.SubjectTypeNikkiUser)
// 	EntitlementSubjectTypeNikkiGroup = EntitlementSubjectType(entEntitlement.SubjectTypeNikkiGroup)
// 	EntitlementSubjectTypeNikkiRole  = EntitlementSubjectType(entEntitlement.SubjectTypeNikkiRole)
// 	EntitlementSubjectTypeCustom     = EntitlementSubjectType(entEntitlement.SubjectTypeCustom)
// )

// func (this EntitlementSubjectType) String() string {
// 	return string(this)
// }

// func WrapEntitlementSubjectType(s string) *EntitlementSubjectType {
// 	st := EntitlementSubjectType(s)
// 	return &st
// }

// func WrapEntitlementSubjectTypeEnt(s entEntitlement.SubjectType) *EntitlementSubjectType {
// 	st := EntitlementSubjectType(s)
// 	return &st
// }

// func EntitlementSubjectTypeValidateRule(field **EntitlementSubjectType, isRequired bool) *val.FieldRules {
// 	return val.Field(field,
// 		val.NotNilWhen(isRequired),
// 		val.When(field != nil,
// 			val.NotEmpty,
// 			val.OneOf(EntitlementSubjectTypeNikkiUser, EntitlementSubjectTypeNikkiGroup, EntitlementSubjectTypeNikkiRole, EntitlementSubjectTypeCustom),
// 		),
// 	)
// }

// func EntitlementSubjectRefValidateRule(subjectType **EntitlementSubjectType, subjectRef **string, isRequired bool) *val.FieldRules {
// 	switch **subjectType {
// 	case EntitlementSubjectTypeCustom:
// 		return val.Field(subjectRef,
// 			val.NotNilWhen(isRequired),
// 			val.NotEmpty,
// 			val.Length(1, model.MODEL_RULE_LONG_NAME_LENGTH),
// 		)
// 	default:
// 		return model.IdPtrValidateRule(subjectRef, isRequired)
// 	}
// }

func EntitlementScopeRefValidateRule(field **string) *val.FieldRules {
	return val.Field(field,
		val.When(*field != nil,
			val.NotEmpty,
			val.Length(model.MODEL_RULE_ULID_LENGTH, model.MODEL_RULE_ULID_LENGTH),
		),
	)
}

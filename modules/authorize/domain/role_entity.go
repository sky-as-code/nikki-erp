package domain

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	entRole "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/role"
)

type Role struct {
	model.ModelBase
	model.AuditableBase

	Name                 *string        `json:"name,omitempty"`
	Description          *string        `json:"description,omitempty"`
	OwnerType            *RoleOwnerType `json:"ownerType,omitempty"`
	OwnerRef             *model.Id      `json:"ownerRef,omitempty"`
	IsRequestable        *bool          `json:"isRequestable,omitempty"`
	IsRequiredAttachment *bool          `json:"isRequiredAttachment,omitempty"`
	IsRequiredComment    *bool          `json:"isRequiredComment,omitempty"`
	CreatedBy            *model.Id      `json:"createdBy,omitempty"`

	Entitlements []*Entitlement `json:"entitlements,omitempty"`
}

func (this *Role) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
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
		val.Field(&this.IsRequestable,
			val.NotNilWhen(!forEdit),
		),
		val.Field(&this.IsRequiredAttachment,
			val.NotNilWhen(!forEdit),
		),
		val.Field(&this.IsRequiredComment,
			val.NotNilWhen(!forEdit),
		),
		RoleOwnerTypeValidateRule(&this.OwnerType, !forEdit),
		model.IdPtrValidateRule(&this.OwnerRef, !forEdit),
		model.IdPtrValidateRule(&this.CreatedBy, !forEdit),
	}
	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)

	return val.ApiBased.ValidateStruct(this, rules...)
}

type RoleOwnerType string

const (
	RoleOwnerTypeUser  = RoleOwnerType(entRole.OwnerTypeUser)
	RoleOwnerTypeGroup = RoleOwnerType(entRole.OwnerTypeGroup)
)

func (this RoleOwnerType) String() string {
	return string(this)
}

func WrapRoleOwnerType(s string) *RoleOwnerType {
	ot := RoleOwnerType(s)
	return &ot
}

func WrapRoleOwnerTypeEnt(s entRole.OwnerType) *RoleOwnerType {
	ot := RoleOwnerType(s)
	return &ot
}

func RoleOwnerTypeValidateRule(field **RoleOwnerType, isRequired bool) *val.FieldRules {
	return val.Field(field,
		val.NotNilWhen(isRequired),
		val.When(*field != nil,
			val.NotEmpty,
			val.OneOf(RoleOwnerTypeUser, RoleOwnerTypeGroup),
		),
	)
}

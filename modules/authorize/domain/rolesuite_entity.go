package domain

import (
	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	entRoleSuite "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/rolesuite"
)

type RoleSuite struct {
	model.ModelBase
	model.AuditableBase

	Name                 *string             `json:"displayName,omitempty"`
	Description          *string             `json:"description,omitempty"`
	OwnerType            *RoleSuiteOwnerType `json:"ownerType,omitempty"`
	OwnerRef             *model.Id           `json:"ownerRef,omitempty"`
	IsRequestable        *bool               `json:"isRequestable,omitempty"`
	IsRequiredAttachment *bool               `json:"isRequiredAttachment,omitempty"`
	IsRequiredComment    *bool               `json:"isRequiredComment,omitempty"`
	CreatedBy            *model.Id           `json:"createdBy,omitempty"`
	OrgId                *model.Id           `json:"orgId,omitempty"`

	Roles []Role `json:"roles,omitempty" model:"-"` // TODO: Handle copy
}

func (this *RoleSuite) Validate(forEdit bool) ft.ValidationErrors {
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
		RoleSuiteOwnerTypeValidateRule(&this.OwnerType, !forEdit),
		model.IdPtrValidateRule(&this.OwnerRef, !forEdit),
		model.IdPtrValidateRule(&this.CreatedBy, !forEdit),
	}
	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)

	return val.ApiBased.ValidateStruct(this, rules...)
}

type RoleSuiteOwnerType string

const (
	RoleSuiteOwnerTypeUser  = RoleSuiteOwnerType(entRoleSuite.OwnerTypeUser)
	RoleSuiteOwnerTypeGroup = RoleSuiteOwnerType(entRoleSuite.OwnerTypeGroup)
)

func (this RoleSuiteOwnerType) Validate() error {
	switch this {
	case RoleSuiteOwnerTypeUser, RoleSuiteOwnerTypeGroup:
		return nil
	default:
		return errors.Errorf("invalid owner type value: %s", this)
	}
}

func (this RoleSuiteOwnerType) String() string {
	return string(this)
}

func WrapRoleSuiteOwnerType(s string) *RoleSuiteOwnerType {
	ot := RoleSuiteOwnerType(s)
	return &ot
}

func WrapRoleSuiteOwnerTypeEnt(s entRoleSuite.OwnerType) *RoleSuiteOwnerType {
	ot := RoleSuiteOwnerType(s)
	return &ot
}

func RoleSuiteOwnerTypeValidateRule(field **RoleSuiteOwnerType, isRequired bool) *val.FieldRules {
	return val.Field(field,
		val.NotNilWhen(isRequired),
		val.When(*field != nil,
			val.NotEmpty,
			val.OneOf(RoleSuiteOwnerTypeUser, RoleSuiteOwnerTypeGroup),
		),
	)
}

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

	Name                 *string   `json:"displayName,omitempty"`
	Description          *string   `json:"description,omitempty"`
	Etag                 *string   `json:"etag,omitempty"`
	OwnerType            *string   `json:"ownerType,omitempty"`
	OwnerRef             *model.Id `json:"ownerRef,omitempty"`
	IsRequestable        *bool     `json:"isRequestable,omitempty"`
	IsRequiredAttachment *bool     `json:"isRequiredAttachment,omitempty"`
	IsRequiredComment    *bool     `json:"isRequiredComment,omitempty"`

	Roles []Role `json:"roles,omitempty"`
}

func (this *RoleSuite) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		val.Field(&this.Name,
			val.RequiredWhen(!forEdit),
			val.Length(1, model.MODEL_RULE_SHORT_NAME_LENGTH),
		),
		val.Field(&this.Description,
			val.Required,
			val.Length(1, model.MODEL_RULE_DESC_LENGTH),
		),
		RoleSuiteOwnerTypeValidateRule(&this.OwnerType, !forEdit),
		model.IdValidateRule(&this.OwnerRef, !forEdit),
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

func RoleSuiteOwnerTypeValidateRule(field any, isRequired bool) *val.FieldRules {
	return val.Field(field,
		val.RequiredWhen(isRequired),
		val.OneOf(RoleSuiteOwnerTypeUser, RoleSuiteOwnerTypeGroup),
	)
}

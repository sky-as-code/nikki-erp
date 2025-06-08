package domain

import (
	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	entRole "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/role"
)

type Role struct {
	model.ModelBase
	model.AuditableBase

	Name                 *string        `json:"displayName,omitempty"`
	Description          *string        `json:"description,omitempty"`
	Etag                 *string        `json:"etag,omitempty"`
	OwnerType            *RoleOwnerType `json:"ownerType,omitempty"`
	OwnerRef             *model.Id      `json:"ownerRef,omitempty"`
	IsRequestable        *bool          `json:"isRequestable,omitempty"`
	IsRequiredAttachment *bool          `json:"isRequiredAttachment,omitempty"`
	IsRequiredComment    *bool          `json:"isRequiredComment,omitempty"`

	Entitlements []Entitlement `json:"entitlements,omitempty"`
}

func (this *Role) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		val.Field(&this.Name,
			val.RequiredWhen(!forEdit),
			val.Length(1, model.MODEL_RULE_SHORT_NAME_LENGTH),
		),
		val.Field(&this.Description,
			val.Length(1, model.MODEL_RULE_DESC_LENGTH),
		),
		RoleOwnerTypeValidateRule(&this.OwnerType, !forEdit),
		model.IdValidateRule(&this.OwnerRef, !forEdit),
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

func (this RoleOwnerType) Validate() error {
	switch this {
	case RoleOwnerTypeUser, RoleOwnerTypeGroup:
		return nil
	default:
		return errors.Errorf("invalid owner type value: %s", this)
	}
}

func (this RoleOwnerType) String() string {
	return string(this)
}

func WrapRoleOwnerType(s string) *RoleOwnerType {
	ot := RoleOwnerType(s)
	return &ot
}

func RoleOwnerTypeValidateRule(field any, isRequired bool) *val.FieldRules {
	return val.Field(field,
		val.RequiredWhen(isRequired),
		val.OneOf(RoleOwnerTypeUser, RoleOwnerTypeGroup),
	)
}

package domain

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"go.bryk.io/pkg/errors"
)

type Role struct {
	model.ModelBase
	model.AuditableBase

	DisplayName          *string        `json:"displayName,omitempty"`
	Description          *string        `json:"description,omitempty"`
	OwnerType            *RoleOwnerType `json:"ownerType,omitempty"`
	OwnerId              *model.Id      `json:"ownerId,omitempty"`
	IsRequestable        *bool          `json:"isRequestable,omitempty"`
	IsRequiredAttachment *bool          `json:"isRequiredAttachment,omitempty"`
	IsRequiredComment    *bool          `json:"isRequiredComment,omitempty"`

	Entitlements []Entitlement `json:"entitlements,omitempty"`
}

func (this *Role) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		val.Field(&this.DisplayName,
			val.RequiredWhen(!forEdit),
			val.Length(1, 200),
		),
		val.Field(&this.Description,
			val.Length(1, 3000),
		),
		RoleOwnerTypeValidateRule(&this.OwnerType, !forEdit),
		model.IdValidateRule(&this.OwnerId, !forEdit),
	}
	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)

	return val.ApiBased.ValidateStruct(this, rules...)
}

type RoleOwnerType string

const (
	RoleOwnerTypeUser  RoleOwnerType = "user"
	RoleOwnerTypeGroup RoleOwnerType = "group"
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

package domain

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
)

type User struct {
	model.ModelBase
	model.AuditableBase

	AvatarUrl   *string     `json:"avatarUrl"`
	DisplayName *string     `json:"displayName"`
	Email       *string     `json:"email"`
	HierarchyId *model.Id   `json:"hierarchyId"`
	Status      *UserStatus `json:"status,omitempty"`

	Groups      []Group          `json:"groups,omitempty" model:"-"` // TODO: Handle copy
	Hierarchies []HierarchyLevel `json:"hierarchies,omitempty" model:"-"`
	Orgs        []Organization   `json:"orgs,omitempty" model:"-"`
}

func (this *User) SetDefaults() {
	this.ModelBase.SetDefaults()
}

func (this *User) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		val.Field(&this.AvatarUrl,
			val.When(this.AvatarUrl != nil,
				val.Length(1, model.MODEL_RULE_URL_LENGTH),
				val.IsUrl,
			),
		),
		val.Field(&this.DisplayName,
			val.NotNilWhen(!forEdit),
			val.When(this.DisplayName != nil,
				val.NotEmpty,
				val.Length(1, model.MODEL_RULE_LONG_NAME_LENGTH),
			),
		),
		val.Field(&this.Email,
			val.NotNilWhen(!forEdit),
			val.When(this.Email != nil,
				val.NotEmpty,
				val.IsEmail,
				val.Length(5, model.MODEL_RULE_EMAIL_LENGTH),
			),
		),
		UserStatusValidateRule(&this.Status),
		model.IdPtrValidateRule(&this.HierarchyId, false),
	}
	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)

	return val.ApiBased.ValidateStruct(this, rules...)
}

type UserStatus string

const (
	UserStatusActive   = UserStatus("active")
	UserStatusArchived = UserStatus("archived")
	UserStatusLocked   = UserStatus("locked")
)

func (this UserStatus) String() string {
	return string(this)
}

func WrapUserStatus(s string) *UserStatus {
	st := UserStatus(s)
	return &st
}

func UserStatusValidateRule(field **UserStatus) *val.FieldRules {
	return val.Field(field,
		val.When(*field != nil,
			val.NotEmpty,
			val.OneOf(UserStatusActive, UserStatusArchived, UserStatusLocked),
		),
	)
}

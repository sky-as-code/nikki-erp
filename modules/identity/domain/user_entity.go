package domain

import (
	"time"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/safe"
	util "github.com/sky-as-code/nikki-erp/common/util"
	val "github.com/sky-as-code/nikki-erp/common/validator"
)

type User struct {
	model.ModelBase
	model.AuditableBase

	AvatarUrl          *string     `json:"avatarUrl"`
	DisplayName        *string     `json:"displayName"`
	Email              *string     `json:"email"`
	FailedLoginUsers   *int        `json:"failedLoginUsers"`
	HierarchyId        *model.Id   `json:"hierarchyId"`
	LastLoginAt        *time.Time  `json:"lastLoginAt"`
	LockedUntil        *time.Time  `json:"lockedUntil"`
	MustChangePassword *bool       `json:"mustChangePassword"`
	PasswordRaw        *string     `json:"passwordRaw"`
	PasswordChangedAt  *time.Time  `json:"passwordChangedAt"`
	PasswordHash       *string     `json:"passwordHash"`
	Status             *UserStatus `json:"status,omitempty"`

	Groups      []Group          `json:"groups,omitempty" model:"-"` // TODO: Handle copy
	Hierarchies []HierarchyLevel `json:"hierarchies,omitempty" model:"-"`
	Orgs        []Organization   `json:"orgs,omitempty" model:"-"`
}

func (this *User) SetDefaults() {
	this.ModelBase.SetDefaults()

	now := time.Now()

	if !util.IsEmptyStr(this.PasswordRaw) {
		this.PasswordChangedAt = &now
	}

	if this.FailedLoginUsers == nil || *this.FailedLoginUsers < 0 {
		this.FailedLoginUsers = util.ToPtr(0)
	}

	safe.SetDefaultValue(&this.MustChangePassword, true)
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
		val.Field(&this.FailedLoginUsers,
			val.Min(0),
			val.Max(model.MODEL_RULE_MAX_INT16),
		),
		val.Field(&this.PasswordRaw,
			val.NotNilWhen(!forEdit),
			val.When(this.PasswordRaw != nil,
				val.NotEmpty,
				val.Length(model.MODEL_RULE_PASSWORD_MIN_LENGTH, model.MODEL_RULE_PASSWORD_MAX_LENGTH),
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

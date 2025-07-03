package domain

import (
	"time"

	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/safe"
	util "github.com/sky-as-code/nikki-erp/common/util"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	entUser "github.com/sky-as-code/nikki-erp/modules/identity/infra/ent/user"
)

type User struct {
	model.ModelBase
	model.AuditableBase
	// model.OrgBase

	AvatarUrl           *string     `json:"avatarUrl,omitempty"`
	DisplayName         *string     `json:"displayName,omitempty"`
	Email               *string     `json:"email,omitempty"`
	FailedLoginAttempts *int        `json:"failedLoginAttempts,omitempty"`
	LastLoginAt         *time.Time  `json:"lastLoginAt,omitempty"`
	LockedUntil         *time.Time  `json:"lockedUntil,omitempty"`
	MustChangePassword  *bool       `json:"mustChangePassword,omitempty"`
	PasswordRaw         *string     `json:"passwordRaw,omitempty"`
	PasswordChangedAt   *time.Time  `json:"passwordChangedAt,omitempty"`
	PasswordHash        *string     `json:"passwordHash,omitempty"`
	Status              *UserStatus `json:"status,omitempty"`

	Groups      []Group          `json:"groups,omitempty"`
	Hierarchies []HierarchyLevel `json:"hierarchies,omitempty"`
	Orgs        []Organization   `json:"orgs,omitempty"`
}

func (this *User) SetDefaults() error {
	err := this.ModelBase.SetDefaults()
	if err != nil {
		return err
	}

	safe.SetDefaultValue(&this.Status, UserStatusInactive)

	now := time.Now()

	if !util.IsEmptyStr(this.PasswordRaw) {
		this.PasswordChangedAt = &now
	}

	if this.FailedLoginAttempts == nil || *this.FailedLoginAttempts < 0 {
		this.FailedLoginAttempts = util.ToPtr(0)
	}

	safe.SetDefaultValue(&this.MustChangePassword, true)

	return nil
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
		val.Field(&this.FailedLoginAttempts,
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
	}
	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)

	return val.ApiBased.ValidateStruct(this, rules...)
}

type UserStatus entUser.Status

const (
	UserStatusActive   = UserStatus(entUser.StatusActive)
	UserStatusInactive = UserStatus(entUser.StatusInactive)
	UserStatusLocked   = UserStatus(entUser.StatusLocked)
)

func (this UserStatus) Validate() error {
	switch this {
	case UserStatusActive, UserStatusInactive, UserStatusLocked:
		return nil
	default:
		return errors.Errorf("invalid status value: %s", this)
	}
}

func (this UserStatus) String() string {
	return string(this)
}

func WrapUserStatus(s string) *UserStatus {
	st := UserStatus(s)
	return &st
}

func WrapUserStatusEnt(s entUser.Status) *UserStatus {
	st := UserStatus(s)
	return &st
}

func UserStatusValidateRule(field **UserStatus) *val.FieldRules {
	return val.Field(field,
		val.When(*field != nil,
			val.NotEmpty,
			val.OneOf(UserStatusActive, UserStatusInactive, UserStatusLocked),
		),
	)
}

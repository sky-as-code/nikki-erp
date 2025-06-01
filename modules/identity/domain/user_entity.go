package domain

import (
	"time"

	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
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

	Groups []Group        `json:"groups,omitempty"`
	Orgs   []Organization `json:"orgs,omitempty"`
}

func (this *User) SetDefaults() error {
	err := this.ModelBase.SetDefaults()
	if err != nil {
		return err
	}

	util.SetDefaultValue(this.Status, UserStatusInactive)

	now := time.Now()

	if !util.IsEmptyStr(this.PasswordRaw) {
		this.PasswordChangedAt = &now
	}

	if this.FailedLoginAttempts == nil || *this.FailedLoginAttempts < 0 {
		*this.FailedLoginAttempts = 0
	}

	util.SetDefaultValue(this.MustChangePassword, true)

	return nil
}

func (this *User) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		val.Field(&this.AvatarUrl, val.When(this.AvatarUrl != nil,
			val.Length(1, 255),
			val.IsUrl,
		)),
		val.Field(&this.DisplayName,
			val.RequiredWhen(!forEdit),
			val.Length(1, 50),
		),
		val.Field(&this.Email,
			val.RequiredWhen(!forEdit),
			val.IsEmail,
			val.Length(5, 100),
		),
		val.Field(&this.PasswordRaw,
			val.RequiredWhen(!forEdit),
			val.Length(8, 100),
		),
		UserStatusValidateRule(&this.Status),
	}
	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)
	// rules = append(rules, this.OrgBase.ValidateRules(forEdit)...)

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

func UserStatusValidateRule(field any) *val.FieldRules {
	return val.Field(field,
		val.OneOf(UserStatusActive, UserStatusInactive, UserStatusLocked),
	)
}

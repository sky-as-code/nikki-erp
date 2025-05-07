package domain

import (
	"time"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/model"
	util "github.com/sky-as-code/nikki-erp/common/util"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	ent "github.com/sky-as-code/nikki-erp/modules/identity/infra/ent/user"
)

type User struct {
	model.ModelBase
	model.AuditableBase
	// model.OrgBase

	AvatarUrl           *string
	DisplayName         *string
	Email               *string
	FailedLoginAttempts *int
	LastLoginAt         *time.Time
	LockedUntil         *time.Time
	MustChangePassword  *bool
	PasswordRaw         *string
	PasswordChangedAt   *time.Time
	PasswordHash        *string
	Status              *UserStatus

	Groups []*Group
	Orgs   []*Organization
}

func (this *User) SetDefaults() {
	this.ModelBase.SetDefaults()

	util.SetDefaultValue(this.Status, UserDefaultStatus)

	now := time.Now()

	if !util.IsEmptyStr(this.PasswordRaw) {
		this.PasswordChangedAt = &now
	}

	if this.FailedLoginAttempts == nil || *this.FailedLoginAttempts < 0 {
		*this.FailedLoginAttempts = 0
	}

	util.SetDefaultValue(this.MustChangePassword, true)
}

func (this *User) Validate(forEdit bool) error {
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

type UserStatus ent.Status

const UserDefaultStatus = UserStatus(ent.DefaultStatus)

const (
	UserStatusActive   = UserStatus(ent.StatusActive)
	UserStatusInactive = UserStatus(ent.StatusInactive)
	UserStatusLocked   = UserStatus(ent.StatusLocked)
)

func (this UserStatus) Validate() error {
	switch this {
	case UserStatusActive, UserStatusInactive, UserStatusLocked:
		return nil
	default:
		return errors.Errorf("invalid status value: %s", this)
	}
}

func WrapUserStatus(s string) *UserStatus {
	st := UserStatus(s)
	return &st
}

func WrapUserStatusEnt(s ent.Status) *UserStatus {
	st := UserStatus(s)
	return &st
}

func UserStatusValidateRule(field any) *val.FieldRules {
	return val.Field(field,
		val.OneOf(UserStatusActive, UserStatusInactive, UserStatusLocked),
	)
}

package domain

import (
	"time"

	"github.com/sky-as-code/nikki-erp/common/array"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/safe"
	util "github.com/sky-as-code/nikki-erp/common/util"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/enum"
)

type User struct {
	model.ModelBase
	model.AuditableBase

	AvatarUrl           *string    `json:"avatarUrl,omitempty"`
	DisplayName         *string    `json:"displayName,omitempty"`
	Email               *string    `json:"email,omitempty"`
	FailedLoginAttempts *int       `json:"failedLoginAttempts,omitempty"`
	HierarchyId         *model.Id  `json:"hierarchyId,omitempty"`
	LastLoginAt         *time.Time `json:"lastLoginAt,omitempty"`
	LockedUntil         *time.Time `json:"lockedUntil,omitempty"`
	MustChangePassword  *bool      `json:"mustChangePassword,omitempty"`
	PasswordRaw         *string    `json:"passwordRaw,omitempty"`
	PasswordChangedAt   *time.Time `json:"passwordChangedAt,omitempty"`
	PasswordHash        *string    `json:"passwordHash,omitempty"`
	StatusId            *model.Id  `json:"statusId,omitempty"`
	StatusValue         *string    `json:"statusValue,omitempty"`

	Groups      []Group          `json:"groups,omitempty"`
	Hierarchies []HierarchyLevel `json:"hierarchies,omitempty"`
	Orgs        []Organization   `json:"orgs,omitempty"`
	Status      *UserStatus      `json:"status,omitempty"`
}

func (this *User) SetDefaults() error {
	err := this.ModelBase.SetDefaults()
	if err != nil {
		return err
	}

	// safe.SetDefaultValue(&this.Status, UserStatusInactive)

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
		model.IdPtrValidateRule(&this.HierarchyId, false),
		model.IdPtrValidateRule(&this.StatusId, false),
	}
	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)

	return val.ApiBased.ValidateStruct(this, rules...)
}

type UserStatus struct {
	enum.Enum
}

func WrapUserStatus(status *enum.Enum) *UserStatus {
	return &UserStatus{
		Enum: *status,
	}
}

func WrapUserStatuses(statuses []enum.Enum) []UserStatus {
	return array.Map(statuses, func(status enum.Enum) UserStatus {
		return *WrapUserStatus(&status)
	})
}

const (
	UserStatusActive   = "active"
	UserStatusArchived = "archived"
	UserStatusLocked   = "locked"
	UserStatusEnumType = "ident_user_status"
)

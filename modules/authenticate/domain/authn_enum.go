package domain

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	enum "github.com/sky-as-code/nikki-erp/modules/core/enum/interfaces"
)

type AuthNEnum = enum.Enum

func WrapAuthNEnum(enum *enum.Enum) *AuthNEnum {
	return (*AuthNEnum)(enum)
}

func WrapAuthNEnums(enums []enum.Enum) []AuthNEnum {
	return array.Map(enums, func(enum enum.Enum) AuthNEnum {
		return *WrapAuthNEnum(&enum)
	})
}

const (
	AuthNAttemptStatusEnumType = "authn_attempt_status"

	AuthNAttemptStatusSuccess = "success"
	AuthNAttemptStatusFailed  = "failed"
	AuthNAttemptStatusPending = "pending"
)

type AttemptStatus string

const (
	AttemptStatusSuccess = AttemptStatus(AuthNAttemptStatusSuccess)
	AttemptStatusFailed  = AttemptStatus(AuthNAttemptStatusFailed)
	AttemptStatusPending = AttemptStatus(AuthNAttemptStatusPending)
)

func (this AttemptStatus) String() string {
	return string(this)
}

func AttemptStatusValidateRule(field **AttemptStatus) *val.FieldRules {
	return val.Field(field,
		val.When(*field != nil,
			val.NotEmpty,
			val.OneOf(AttemptStatusPending, AttemptStatusSuccess, AttemptStatusFailed),
		),
	)
}

type SettingSubjectType string

const (
	SettingSubjectTypeDomain = SettingSubjectType("domain")
	SettingSubjectTypeOrg    = SettingSubjectType("org")
	SettingSubjectTypeUser   = SettingSubjectType("user")
	SettingSubjectTypeCustom = SettingSubjectType("custom")
)

func (this SettingSubjectType) String() string {
	return string(this)
}

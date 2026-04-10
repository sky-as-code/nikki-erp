package domain

import (
	"github.com/sky-as-code/nikki-erp/common/array"
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

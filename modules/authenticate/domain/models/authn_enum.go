package models

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

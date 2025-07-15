package domain

import val "github.com/sky-as-code/nikki-erp/common/validator"

type SubjectType string

const (
	SubjectTypeUser   = SubjectType("user")
	SubjectTypeCustom = SubjectType("custom")
)

func (this SubjectType) String() string {
	return string(this)
}

func WrapSubjectType(s string) *SubjectType {
	st := SubjectType(s)
	return &st
}

func SubjectTypeValidateRule(field **SubjectType) *val.FieldRules {
	return val.Field(field,
		val.When(*field != nil,
			val.NotEmpty,
			val.OneOf(SubjectTypeUser, SubjectTypeCustom),
		),
	)
}

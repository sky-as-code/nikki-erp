package domain

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
)

type MethodSetting struct {
	model.ModelBase

	Method           *string             `json:"method"`
	Order            *int                `json:"order"`
	MaxFailures      *int                `json:"maxFailures"`
	LockDurationSec  *int                `json:"lockDurationSec"`
	SubjectType      *SettingSubjectType `json:"subjectType"`
	SubjectRef       *model.Id           `json:"subjectRef"`
	SubjectSourceRef *string             `json:"subjectSourceRef"`
}

func (this *MethodSetting) SetDefaults() {
	this.ModelBase.SetDefaults()
}

func (this *MethodSetting) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		val.Field(&this.Method,
			val.NotNilWhen(!forEdit),
			val.When(this.Method != nil,
				val.NotEmpty,
			),
		),
		val.Field(&this.Order,
			val.NotNilWhen(!forEdit),
			val.When(this.Order != nil,
				val.Min(1),
				val.Min(5),
			),
		),
		val.Field(&this.MaxFailures,
			val.NotNilWhen(!forEdit),
			val.When(this.MaxFailures != nil,
				val.Min(1),
				val.Max(model.MODEL_RULE_MAX_INT16),
			),
		),

		val.Field(&this.LockDurationSec,
			val.NotNilWhen(!forEdit),
			val.When(this.LockDurationSec != nil,
				val.Min(0),
				val.Max(model.MODEL_RULE_MAX_INT64),
			),
		),

		model.IdPtrValidateRule(&this.SubjectRef, true),
		SettingSubjectTypeValidateRule(&this.SubjectType),
	}
	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)

	return val.ApiBased.ValidateStruct(this, rules...)
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

func WrapSettingSubjectType(s string) *SettingSubjectType {
	st := SettingSubjectType(s)
	return &st
}

func SettingSubjectTypeValidateRule(field **SettingSubjectType) *val.FieldRules {
	return val.Field(field,
		val.When(*field != nil,
			val.NotEmpty,
			val.OneOf(SettingSubjectTypeDomain, SettingSubjectTypeOrg, SettingSubjectTypeUser, SettingSubjectTypeCustom),
		),
	)
}

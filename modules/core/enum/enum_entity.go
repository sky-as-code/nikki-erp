package enum

import (
	"regexp"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
)

type EnumType = string
type EnumValue = string

type Enum struct {
	model.ModelBase

	Label *model.LangJson `json:"label"`
	Type  *EnumType       `json:"type"`
	Value *EnumValue      `json:"value"`
}

func (this *Enum) Validate(forEdit bool) ft.ValidationErrors {
	rules := this.ValidateRules(forEdit)
	return val.ApiBased.ValidateStruct(this, rules...)
}

func (this *Enum) ValidateRules(forEdit bool) []*val.FieldRules {
	rules := []*val.FieldRules{
		model.LangJsonValidateRule(&this.Label, true, 1, model.MODEL_RULE_TINY_NAME_LENGTH),
		EnumTypeValidateRule(&this.Type, !forEdit),
		EnumValueValidateRule(&this.Value, !forEdit),
	}
	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)

	return rules
}

func EnumTypeValidateRule(field **EnumType, isRequired bool) *val.FieldRules {
	return val.Field(field,
		val.NotNilWhen(isRequired),
		val.When(field != nil,
			val.NotEmpty,
			val.Length(1, model.MODEL_RULE_TINY_NAME_LENGTH),
			val.RegExp(regexp.MustCompile("^[a-zA-Z0-9_]+$")),
		),
	)
}

func EnumValueValidateRule(field **EnumValue, isRequired bool) *val.FieldRules {
	return val.Field(field,
		val.NotNilWhen(isRequired),
		val.When(field != nil,
			val.NotEmpty,
			val.Length(1, model.MODEL_RULE_TINY_NAME_LENGTH),
			val.RegExp(regexp.MustCompile("^[a-zA-Z0-9_]+$")),
		),
	)
}

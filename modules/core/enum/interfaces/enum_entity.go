package interfaces

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

	Label *model.LangJson `json:"label,omitempty"`
	Type  *EnumType       `json:"type,omitempty"`
	Value *EnumValue      `json:"value,omitempty"`
}

func (this *Enum) Validate(forEdit bool) ft.ValidationErrors {
	rules := this.ValidateRules(forEdit)
	return val.ApiBased.ValidateStruct(this, rules...)
}

func (this *Enum) ValidateRules(forEdit bool) []*val.FieldRules {
	rules := []*val.FieldRules{
		model.LangJsonPtrValidateRule(&this.Label, true, 1, model.MODEL_RULE_TINY_NAME_LENGTH),
		EnumTypeValidateRule(&this.Type, !forEdit),
		EnumValueValidateRule(&this.Value, false),
	}
	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)

	return rules
}

func EnumTypeValidateRule(field **EnumType, isRequired bool) *val.FieldRules {
	return val.Field(field,
		val.NotNilWhen(isRequired),
		val.When(*field != nil,
			val.NotEmpty,
			val.Length(1, model.MODEL_RULE_TINY_NAME_LENGTH),
			val.RegExp(regexp.MustCompile("^[a-zA-Z0-9_]+$")),
		),
	)
}

func EnumValueValidateRule(field **EnumValue, isRequired bool) *val.FieldRules {
	return val.Field(field,
		val.NotNilWhen(isRequired),
		val.When(*field != nil,
			val.NotEmpty,
			val.Length(1, model.MODEL_RULE_TINY_NAME_LENGTH),
			val.RegExp(regexp.MustCompile(`^\p{L}[\p{L}\p{N}_]*$`)), // allows all visible Unicode letters, digits and underscore.
		),
	)
}

package interfaces

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
)

type AttributeValue struct {
	model.ModelBase
	model.AuditableBase

	AttributeId  *model.Id       `json:"attributeId,omitempty"`
	ValueText    *model.LangJson `json:"valueText,omitempty"`
	ValueNumber  *float64        `json:"valueNumber,omitempty"`
	ValueBool    *bool           `json:"valueBool,omitempty"`
	ValueRef     *string         `json:"valueRef,omitempty"`
	VariantCount *int            `json:"variantCount,omitempty"`
}

func (this *AttributeValue) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdPtrValidateRule(&this.AttributeId, !forEdit),
		val.Field(&this.ValueText,
			val.When(this.ValueText != nil,
				val.NotEmpty,
				val.Length(1, model.MODEL_RULE_LONG_NAME_LENGTH),
			),
		),
		val.Field(&this.ValueRef,
			val.When(this.ValueRef != nil,
				val.Length(0, model.MODEL_RULE_LONG_NAME_LENGTH),
			),
		),
	}

	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)

	return val.ApiBased.ValidateStruct(this, rules...)
}

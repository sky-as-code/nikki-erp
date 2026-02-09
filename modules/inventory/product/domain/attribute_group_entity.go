package domain

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
)

type AttributeGroup struct {
	model.ModelBase
	model.AuditableBase

	ProductId *string         `json:"productId,omitempty"`
	Name      *model.LangJson `json:"name,omitempty"`
	Index     *int            `json:"index,omitempty"`
}

func (this *AttributeGroup) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		val.Field(&this.ProductId,
			val.NotNilWhen(!forEdit),
			val.When(this.ProductId != nil,
				val.NotEmpty,
			),
		),
		val.Field(&this.Name,
			val.NotNilWhen(!forEdit),
			val.When(this.Name != nil,
				val.NotEmpty,
				val.Length(1, model.MODEL_RULE_LONG_NAME_LENGTH),
			),
		),
		val.Field(&this.Index,
			val.When(this.Index != nil,
				val.Min(0),
			),
		),
	}

	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)

	return val.ApiBased.ValidateStruct(this, rules...)
}

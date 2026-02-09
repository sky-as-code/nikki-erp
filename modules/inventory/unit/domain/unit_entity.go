package domain

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
)

type Unit struct {
	model.ModelBase
	model.AuditableBase

	Name       *model.LangJson `json:"name,omitempty"`
	Symbol     *string         `json:"symbol,omitempty"`
	BaseUnit   *string         `json:"baseUnit,omitempty"`
	Multiplier *int            `json:"multiplier,omitempty"`
	OrgId      *model.Id       `json:"orgId,omitempty"`

	Status     *string `json:"status,omitempty"`
	CategoryId *string `json:"categoryId,omitempty"`
}

func (this *Unit) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		val.Field(&this.Name,
			val.NotNilWhen(!forEdit),
			val.When(this.Name != nil,
				val.NotEmpty,
				val.Length(1, model.MODEL_RULE_LONG_NAME_LENGTH),
			),
		),
		val.Field(&this.Symbol,
			val.NotNilWhen(!forEdit),
			val.When(this.Symbol != nil,
				val.NotEmpty,
				val.Length(1, model.MODEL_RULE_SHORT_NAME_LENGTH),
			),
		),
		val.Field(&this.Multiplier,
			val.When(this.BaseUnit != nil && *this.BaseUnit != "base",
				val.NotNil,
			),
		),
		val.Field(&this.CategoryId,
			val.When(this.CategoryId != nil,
				val.NotEmpty,
			),
		),
		val.Field(&this.BaseUnit,
			val.When(this.BaseUnit != nil,
				val.NotEmpty,
			),
		),
	}

	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)

	return val.ApiBased.ValidateStruct(this, rules...)
}

func (this *Unit) SetDefaults() {
	this.ModelBase.SetDefaults()
}

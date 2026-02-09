package domain

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
)

type ProductCategory struct {
	model.ModelBase
	model.AuditableBase

	ParentId    *model.Id       `json:"parentId,omitempty"`
	Name        *model.LangJson `json:"name,omitempty"`
	Description *model.LangJson `json:"description,omitempty"`
	Path        *string         `json:"path,omitempty"`
	Level       *int            `json:"level,omitempty"`
	SortIndex   *int            `json:"sortIndex,omitempty"`
}

func (this *ProductCategory) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdPtrValidateRule(&this.ParentId, false),
		val.Field(&this.Name,
			val.NotNilWhen(!forEdit),
			val.When(this.Name != nil,
				val.NotEmpty,
				val.Length(1, model.MODEL_RULE_LONG_NAME_LENGTH),
			),
		),
		val.Field(&this.Description,
			val.When(this.Description != nil,
				val.Length(0, model.MODEL_RULE_LONG_NAME_LENGTH),
			),
		),
	}

	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)

	return val.ApiBased.ValidateStruct(this, rules...)
}

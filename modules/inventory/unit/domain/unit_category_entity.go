package domain

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
)

type UnitCategory struct {
	model.ModelBase
	model.AuditableBase

	OrgId        *string         `json:"orgId,omitempty"`
	Name         *model.LangJson `json:"name,omitempty"`
	Description  *model.LangJson `json:"description,omitempty"`
	Status       *string         `json:"status,omitempty"`
	ThumbnailUrl *string         `json:"thumbnailURL,omitempty"`
}

func (this *UnitCategory) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		val.Field(&this.OrgId,
			val.When(this.OrgId != nil,
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
		val.Field(&this.Description,
			val.When(this.Description != nil,
				val.Length(0, model.MODEL_RULE_LONG_NAME_LENGTH),
			),
		),
		val.Field(&this.ThumbnailUrl,
			val.When(this.ThumbnailUrl != nil,
				val.Length(0, model.MODEL_RULE_LONG_NAME_LENGTH),
			),
		),
	}

	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)

	return val.ApiBased.ValidateStruct(this, rules...)
}

func (this *UnitCategory) SetDefaults() {
	this.ModelBase.SetDefaults()
}

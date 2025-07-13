package domain

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
)

type Group struct {
	model.ModelBase
	model.AuditableBase

	Name        *string   `json:"name"`
	Description *string   `json:"description"`
	OrgId       *model.Id `json:"orgId"`

	Org *Organization `json:"organization,omitempty" model:"-"` // TODO: Handle copy
}

func (this *Group) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		val.Field(&this.Name,
			val.NotNilWhen(!forEdit),
			val.When(this.Name != nil,
				val.NotEmpty,
				val.Length(1, model.MODEL_RULE_LONG_NAME_LENGTH),
			),
		),
		val.Field(&this.Description,
			val.Length(0, model.MODEL_RULE_DESC_LENGTH),
		),
		model.IdPtrValidateRule(&this.OrgId, false),
	}
	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)

	return val.ApiBased.ValidateStruct(this, rules...)
}

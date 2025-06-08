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
	Description *string   `json:"description,omitempty"`
	OrgId       *model.Id `json:"orgId,omitempty"`

	Org *Organization `json:"organization,omitempty"`
}

func (this *Group) SetDefaults() error {
	return this.ModelBase.SetDefaults()
}

func (this *Group) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		val.Field(&this.Name,
			val.NotEmptyWhen(!forEdit),
			val.Length(1, 50),
		),
		val.Field(&this.Description, val.When(this.Description != nil,
			val.NotEmpty,
			val.Length(1, 255),
		)),
		model.IdPtrValidateRule(&this.OrgId, false),
	}
	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)

	return val.ApiBased.ValidateStruct(this, rules...)
}

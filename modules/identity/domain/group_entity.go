package domain

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
)

type Group struct {
	model.ModelBase
	model.AuditableBase
	model.OrgBase

	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`

	ParentId *model.Id `json:"parentId,omitempty"`
}

func (this *Group) SetDefaults() error {
	return this.ModelBase.SetDefaults()
}

func (this *Group) Validate(forEdit bool) error {
	rules := []*val.FieldRules{
		val.Field(&this.Name,
			val.RequiredWhen(!forEdit),
			val.Length(1, 50),
		),
		val.Field(&this.Description,
			val.Length(0, 255),
		),
		model.IdValidateRule(&this.ParentId, !forEdit),
	}
	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)
	rules = append(rules, this.OrgBase.ValidateRules(forEdit)...)

	return val.ApiBased.ValidateStruct(this, rules...)
}

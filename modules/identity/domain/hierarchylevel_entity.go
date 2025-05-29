package domain

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
)

type HierarchyLevel struct {
	model.ModelBase
	model.AuditableBase
	model.OrgBase

	Name     *string
	ParentId *model.Id
}

func (this *HierarchyLevel) SetDefaults() error {
	return this.ModelBase.SetDefaults()
}

func (this *HierarchyLevel) Validate(forEdit bool) error {
	rules := []*val.FieldRules{
		val.Field(&this.Name,
			val.RequiredWhen(!forEdit),
			val.Length(1, 50),
		),
		model.IdValidateRule(&this.ParentId, false),
	}
	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)
	rules = append(rules, this.OrgBase.ValidateRules(forEdit)...)
	return val.ApiBased.ValidateStruct(this, rules...)
}

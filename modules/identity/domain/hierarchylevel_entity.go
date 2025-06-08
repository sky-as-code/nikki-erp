package domain

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
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

func (this *HierarchyLevel) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		val.Field(&this.Name,
			val.NotEmptyWhen(!forEdit),
			val.Length(1, 50),
		),
		model.IdValidateRule(&this.ParentId, false),
	}
	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)
	rules = append(rules, this.OrgBase.ValidateRules(forEdit)...)
	return val.ApiBased.ValidateStruct(this, rules...)
}

package domain

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
)

type HierarchyLevel struct {
	model.ModelBase
	model.AuditableBase

	Name     *string
	ParentId *model.Id
	OrgId    *model.Id
	ScopeRef *model.Id `json:"scopeRef,omitempty" model:"-"`

	Org      *Organization    `json:"org,omitempty" model:"-"`
	Parent   *HierarchyLevel  `json:"parent,omitempty" model:"-"`
	Children []HierarchyLevel `json:"children,omitempty" model:"-"`
}

func (this *HierarchyLevel) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		val.Field(&this.Name,
			val.NotEmptyWhen(!forEdit),
			val.Length(1, 50),
		),
		model.IdPtrValidateRule(&this.ParentId, false),
		model.IdPtrValidateRule(&this.OrgId, !forEdit),
	}
	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)
	return val.ApiBased.ValidateStruct(this, rules...)
}

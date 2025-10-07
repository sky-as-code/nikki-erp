package domain

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
)

type Entitlement struct {
	model.ModelBase
	model.AuditableBase

	ActionId    *model.Id `json:"actionId,omitempty"`
	ActionExpr  *string   `json:"actionExpr,omitempty"`
	Description *string   `json:"description,omitempty"`
	Name        *string   `json:"name,omitempty"`
	ResourceId  *model.Id `json:"resourceId,omitempty"`
	ScopeRef    *string   `json:"scopeRef,omitempty"`
	CreatedBy   *string   `json:"createdBy,omitempty"`
	OrgId       *model.Id `json:"orgId,omitempty"`

	Action   *Action   `json:"action,omitempty" model:"-"` // TODO: Handle copy
	Resource *Resource `json:"resource,omitempty" model:"-"`
}

func (this *Entitlement) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdPtrValidateRule(&this.ActionId, !forEdit),
		val.Field(&this.ActionExpr,
			val.NotNilWhen(!forEdit),
			val.When(this.ActionExpr != nil,
				val.NotEmpty,
			),
		),
		val.Field(&this.Name,
			val.NotNilWhen(!forEdit),
			val.When(this.Name != nil,
				val.NotEmpty,
				val.Length(1, model.MODEL_RULE_SHORT_NAME_LENGTH),
			),
		),
		val.Field(&this.Description,
			val.When(this.Description != nil,
				val.NotEmpty,
				val.Length(1, model.MODEL_RULE_DESC_LENGTH),
			),
		),
		// EntitlementScopeRefValidateRule(&this.ScopeRef),
		model.IdPtrValidateRule(&this.ResourceId, !forEdit),
		model.IdPtrValidateRule(&this.CreatedBy, !forEdit),
	}
	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)

	return val.ApiBased.ValidateStruct(this, rules...)
}

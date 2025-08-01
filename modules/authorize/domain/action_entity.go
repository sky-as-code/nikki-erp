package domain

import (
	"regexp"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
)

type Action struct {
	model.ModelBase
	model.AuditableBase

	Name        *string   `json:"name,omitempty"`
	ResourceId  *model.Id `json:"resourceId,omitempty"`
	Description *string   `json:"description,omitempty"`
	CreatedBy   *string   `json:"createdBy,omitempty"`

	Resource     *Resource     `json:"resource,omitempty"`
	Entitlements []Entitlement `json:"entitlements"`
}

func (this *Action) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		val.Field(&this.Name,
			val.NotNilWhen(!forEdit),
			val.When(this.Name != nil,
				val.NotEmpty,
				val.RegExp(regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)), // alphanumeric, underscore, dash
				val.Length(1, model.MODEL_RULE_TINY_NAME_LENGTH),
			),
		),
		val.Field(&this.Description,
			val.When(this.Description != nil,
				val.NotEmpty,
				val.Length(1, model.MODEL_RULE_DESC_LENGTH),
			),
		),
		model.IdPtrValidateRule(&this.CreatedBy, !forEdit),
		model.IdPtrValidateRule(&this.ResourceId, !forEdit),
	}
	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)

	return val.ApiBased.ValidateStruct(this, rules...)
}

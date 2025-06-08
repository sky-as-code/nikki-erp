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

	Name       *string   `json:"name,omitempty"`
	ResourceId *model.Id `json:"resourceId,omitempty"`
}

func (this *Action) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		val.Field(&this.Name,
			val.Required,
			val.RegExp(regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)), // alphanumeric, underscore, dash
			val.Length(1, model.MODEL_RULE_TINY_NAME_LENGTH),
		),
		model.IdValidateRule(&this.ResourceId, true),
	}
	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)

	return val.ApiBased.ValidateStruct(this, rules...)
}

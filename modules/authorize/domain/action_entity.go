package domain

import (
	"regexp"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
)

type Action struct {
	model.ModelBase

	Name       *string   `json:"name,omitempty"`
	ResourceId *model.Id `json:"resourceId,omitempty"`
}

func (this *Action) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		val.Field(&this.Name,
			val.Required,
			val.RegExp(regexp.MustCompile(`^[a-zA-Z0-9_\-\s]+$`)), // alphanumeric, underscore, dash and space
			val.Length(1, 50),
		),
		model.IdValidateRule(&this.ResourceId, true),
	}

	return val.ApiBased.ValidateStruct(this, rules...)
}

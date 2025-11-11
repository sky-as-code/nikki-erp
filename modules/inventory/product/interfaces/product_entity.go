package interfaces

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	itVariant "github.com/sky-as-code/nikki-erp/modules/inventory/variant/interfaces"
)

type Product struct {
	model.ModelBase
	model.AuditableBase

	OrgId             *model.Id       `json:"orgId,omitempty" `
	Name              *model.LangJson `json:"name,omitempty" `
	Description       *model.LangJson `json:"description,omitempty" `
	Unit              *model.Id       `json:"unit_id, omitempty" `
	Status            *string         `json:"status,omitempty" `
	DefaultsVariantId *model.Id       `json:"defaultsVariantId,omitempty" `
	ThumbnailUrl      *string         `json:"thumbnailUrl,omitempty" `

	// Relations
	Variants []itVariant.Variant `json:"variants,omitempty" model:"-"`
}

func (this *Product) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		val.Field(&this.Name,
			val.NotNilWhen(!forEdit),
			val.When(this.Name != nil,
				val.NotEmpty,
				val.Length(1, model.MODEL_RULE_LONG_NAME_LENGTH),
			),
		),

		model.IdPtrValidateRule(&this.Unit, false),
	}

	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)

	return val.ApiBased.ValidateStruct(this, rules...)
}

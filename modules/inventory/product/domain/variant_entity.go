package domain

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
)

type Variant struct {
	model.ModelBase
	model.AuditableBase

	ProductId     *model.Id               `json:"productId,omitempty"`
	Name          *model.LangJson         `json:"name,omitempty"`
	Sku           *string                 `json:"sku,omitempty"`
	Barcode       *string                 `json:"barcode,omitempty"`
	ProposedPrice *float64                `json:"proposedPrice,omitempty"`
	Status        *string                 `json:"status,omitempty"`
	Attributes    *map[string]interface{} `json:"attributes,omitempty"`

	AttributeValue []AttributeValue `json:"attributeValue,omitempty"`
}

func (this *Variant) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdPtrValidateRule(&this.ProductId, !forEdit),
		val.Field(&this.Name,
			val.NotNilWhen(!forEdit),
			val.When(this.Name != nil,
				val.NotEmpty,
			),
		),
	}

	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)

	return val.ApiBased.ValidateStruct(this, rules...)
}

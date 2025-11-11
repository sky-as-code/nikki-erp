package interfaces

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
)

type Variant struct {
	model.ModelBase
	model.AuditableBase

	ProductId     *model.Id `json:"productId,omitempty"`
	Sku           *string   `json:"sku,omitempty"`
	Barcode       *string   `json:"barcode,omitempty"`
	ProposedPrice *int      `json:"proposedPrice,omitempty"`
	Status        *string   `json:"status,omitempty"`
}

func (this *Variant) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdPtrValidateRule(&this.ProductId, !forEdit),
		val.Field(&this.Sku,
			val.NotNilWhen(!forEdit),
			val.When(this.Sku != nil,
				val.NotEmpty,
				val.Length(1, model.MODEL_RULE_LONG_NAME_LENGTH),
			),
		),
		val.Field(&this.Barcode,
			val.When(this.Barcode != nil,
				val.Length(0, model.MODEL_RULE_LONG_NAME_LENGTH),
			),
		),
		val.Field(&this.Status,
			val.When(this.Status != nil,
				val.NotEmpty,
			),
		),
	}

	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)

	return val.ApiBased.ValidateStruct(this, rules...)
}

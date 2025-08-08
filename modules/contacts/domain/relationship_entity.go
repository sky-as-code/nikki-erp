package domain

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	enum "github.com/sky-as-code/nikki-erp/modules/core/enum/interfaces"
)

type Relationship struct {
	model.ModelBase
	model.AuditableBase

	Note          *string    `json:"note,omitempty"`
	TargetPartyId *model.Id  `json:"targetPartyId"`
	Type          *enum.Enum `json:"type"`
}

func (this *Relationship) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		val.Field(&this.Type,
			val.NotNilWhen(!forEdit),
			val.When(this.Type != nil,
				val.NotEmpty,
				val.OneOf("employee", "spouse", "parent", "sibling", "emergency", "subsidiary"),
			),
		),

		model.IdPtrValidateRule(&this.TargetPartyId, !forEdit),
	}
	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)

	return val.ApiBased.ValidateStruct(this, rules...)
}

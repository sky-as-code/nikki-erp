package domain

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	enum "github.com/sky-as-code/nikki-erp/modules/core/enum/interfaces"
)

type CommChannel struct {
	model.ModelBase
	model.AuditableBase

	Note      *string        `json:"note,omitempty"`
	PartyId   *model.Id      `json:"partyId"`
	Type      *enum.Enum     `json:"type"`
	Value     *string        `json:"value,omitempty"`
	ValueJson *ValueJsonData `json:"valueJson,omitempty"`

	Party *Party `json:"party,omitempty" model:"-"`
}

func (this *CommChannel) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		val.Field(&this.Type,
			val.NotNilWhen(!forEdit),
			val.When(this.Type != nil,
				val.NotEmpty,
				val.OneOf("Phone", "Zalo", "Facebook", "Email", "Post"),
			),
		),
		val.Field(&this.Value,
			val.When(this.Value != nil,
				val.NotEmpty,
				val.Length(1, 255),
			),
		),
		model.IdPtrValidateRule(&this.PartyId, !forEdit),
	}
	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)

	return val.ApiBased.ValidateStruct(this, rules...)
}

type ValueJsonData struct {
	Data any `json:"data"`
}

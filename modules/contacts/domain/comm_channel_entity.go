package domain

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
)

type ValueJsonData = model.LangJson
type CommChannel struct {
	model.ModelBase
	model.AuditableBase

	Note      *string        `json:"note,omitempty"`
	OrgId     *model.Id      `json:"orgId"`
	PartyId   *model.Id      `json:"partyId"`
	Type      *string        `json:"type"`
	Value     *string        `json:"value,omitempty"`
	ValueJson *ValueJsonData `json:"valueJson,omitempty"`

	Party *Party `json:"party,omitempty" model:"-"`
}

func (this *CommChannel) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		val.Field(&this.Type,
			val.NotEmpty,
			val.Length(1, 100),
			val.OneOf(TypePhone, TypeZalo, TypeFacebook, TypeEmail, TypePost),
		),

		val.Field(&this.Value,
			val.When(this.Value != nil,
				val.NotEmpty,
				val.Length(1, 255),
			),
		),
		model.IdPtrValidateRule(&this.OrgId, !forEdit),
		model.IdPtrValidateRule(&this.PartyId, !forEdit),
	}
	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)

	return val.ApiBased.ValidateStruct(this, rules...)
}

const (
	TypePhone    = "phone"
	TypeZalo     = "zalo"
	TypeFacebook = "facebook"
	TypeEmail    = "email"
	TypePost     = "post"
)

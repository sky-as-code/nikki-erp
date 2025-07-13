package domain

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
)

type Organization struct {
	model.ModelBase
	model.AuditableBase

	Address     *string     `json:"address"`
	DisplayName *string     `json:"displayName"`
	LegalName   *string     `json:"legalName"`
	PhoneNumber *string     `json:"phoneNumber"`
	Slug        *model.Slug `json:"slug"`
	StatusId    *model.Id   `json:"statusId"`
	StatusValue *string     `json:"statusValue"`

	Status *IdentityStatus `json:"status,omitempty"`
}

func (this *Organization) SetDefaults() {
	this.ModelBase.SetDefaults()
}

func (this *Organization) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		val.Field(&this.Address,
			val.Length(0, model.MODEL_RULE_LONG_NAME_LENGTH),
		),
		val.Field(&this.DisplayName,
			val.NotNilWhen(!forEdit),
			val.When(this.DisplayName != nil,
				val.NotEmpty,
				val.Length(1, model.MODEL_RULE_SHORT_NAME_LENGTH),
			),
		),
		val.Field(&this.LegalName,
			val.Length(0, model.MODEL_RULE_LONG_NAME_LENGTH),
		),

		model.IdPtrValidateRule(&this.Id, false), // Id is not required but Slug is mandatory in all cases
		model.SlugPtrValidateRule(&this.Slug, true),
		model.EtagPtrValidateRule(&this.Etag, forEdit),
	}
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)

	return val.ApiBased.ValidateStruct(this, rules...)
}

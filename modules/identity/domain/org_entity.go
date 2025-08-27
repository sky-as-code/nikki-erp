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
	Status      *OrgStatus  `json:"status,omitempty"`
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

		OrgStatusValidateRule(&this.Status),
		model.IdPtrValidateRule(&this.Id, false), // Id is not required but Slug is mandatory in all cases
		model.SlugPtrValidateRule(&this.Slug, true),
		model.EtagPtrValidateRule(&this.Etag, forEdit),
	}
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)

	return val.ApiBased.ValidateStruct(this, rules...)
}

type OrgStatus string

const (
	OrgStatusActive   = OrgStatus("active")
	OrgStatusArchived = OrgStatus("archived")
)

func (this OrgStatus) String() string {
	return string(this)
}

func WrapOrgStatus(s string) *OrgStatus {
	st := OrgStatus(s)
	return &st
}

func OrgStatusValidateRule(field **OrgStatus) *val.FieldRules {
	return val.Field(field,
		val.When(*field != nil,
			val.NotEmpty,
			val.OneOf(OrgStatusActive, OrgStatusArchived),
		),
	)
}

package domain

import (
	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/safe"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	ent "github.com/sky-as-code/nikki-erp/modules/identity/infra/ent/organization"
)

type Organization struct {
	model.ModelBase
	model.AuditableBase

	Address     *string     `json:"address"`
	DisplayName *string     `json:"displayName"`
	LegalName   *string     `json:"legalName"`
	PhoneNumber *string     `json:"phoneNumber"`
	Slug        *model.Slug `json:"slug"`
	Status      *OrgStatus  `json:"status"`
}

func (this *Organization) SetDefaults() {
	this.ModelBase.SetDefaults()
	safe.SetDefaultValue(&this.Status, OrgStatusInactive)
}

func (this *Organization) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		val.Field(&this.Address,
			val.NotNilWhen(!forEdit),
			val.When(this.Address != nil,
				val.NotEmpty,
				val.Length(1, model.MODEL_RULE_LONG_NAME_LENGTH),
			),
		),
		val.Field(&this.DisplayName,
			val.NotNilWhen(!forEdit),
			val.When(this.DisplayName != nil,
				val.NotEmpty,
				val.Length(1, model.MODEL_RULE_SHORT_NAME_LENGTH),
			),
		),
		val.Field(&this.LegalName,
			val.NotNilWhen(!forEdit),
			val.When(this.LegalName != nil,
				val.NotEmpty,
				val.Length(1, model.MODEL_RULE_LONG_NAME_LENGTH),
			),
		),

		model.IdPtrValidateRule(&this.Id, false), // Id is not required but Slug is mandatory in all cases
		model.SlugPtrValidateRule(&this.Slug, true),
		model.EtagPtrValidateRule(&this.Etag, forEdit),
	}
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)

	return val.ApiBased.ValidateStruct(this, rules...)
}

type OrgStatus ent.Status

const (
	OrgStatusActive   = OrgStatus(ent.StatusActive)
	OrgStatusInactive = OrgStatus(ent.StatusInactive)
)

func (this OrgStatus) Validate() error {
	switch this {
	case OrgStatusActive, OrgStatusInactive:
		return nil
	default:
		return errors.Errorf("invalid status value: %s", this)
	}
}

func WrapOrgStatus(s string) *OrgStatus {
	st := OrgStatus(s)
	return &st
}

func WrapOrgStatusEnt(s ent.Status) *OrgStatus {
	st := OrgStatus(s)
	return &st
}

func OrgStatusValidateRule(field any) *val.FieldRules {
	return val.Field(field,
		val.OneOf(OrgStatusActive, OrgStatusInactive),
	)
}

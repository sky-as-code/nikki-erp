package domain

import (
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/model"
	util "github.com/sky-as-code/nikki-erp/common/util"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	ent "github.com/sky-as-code/nikki-erp/modules/identity/infra/ent/organization"
)

type Organization struct {
	model.ModelBase
	model.AuditableBase

	DisplayName *string
	Slug        *model.Slug
	Etag        *model.Etag
	Status      *OrgStatus
}

func (this *Organization) SetDefaults() {
	this.ModelBase.SetDefaults()
	util.SetDefaultValue(this.Status, OrgDefaultStatus)
	*this.Etag = model.NewEtag()
}

func (this *Organization) Validate(forEdit bool) error {
	rules := []*val.FieldRules{
		val.Field(&this.DisplayName,
			val.RequiredWhen(!forEdit),
			val.Length(1, 50),
		),
		model.EtagValidateRule(&this.Etag, forEdit),
		model.SlugValidateRule(&this.Slug, !forEdit),
		OrgStatusValidateRule(&this.Status),
	}
	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)

	return val.ApiBased.ValidateStruct(this, rules...)
}

type OrgStatus ent.Status

const OrgDefaultStatus = OrgStatus(ent.DefaultStatus)

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

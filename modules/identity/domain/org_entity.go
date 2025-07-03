package domain

import (
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/safe"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	ent "github.com/sky-as-code/nikki-erp/modules/identity/infra/ent/organization"
)

type Organization struct {
	model.ModelBase
	model.AuditableBase

	Address     *string     `json:"address,omitempty"`
	DisplayName *string     `json:"displayName,omitempty"`
	LegalName   *string     `json:"legalName,omitempty"`
	PhoneNumber *string     `json:"phoneNumber,omitempty"`
	Slug        *model.Slug `json:"slug,omitempty"`
	Status      *OrgStatus  `json:"status,omitempty"`
}

func (this *Organization) SetDefaults() {
	this.ModelBase.SetDefaults()
	safe.SetDefaultValue(&this.Status, OrgStatusInactive)
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

package model

import (
	"time"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/safe"
	val "github.com/sky-as-code/nikki-erp/common/validator"
)

type ModelBase struct {
	Id   *Id   `json:"id,omitempty"`
	Etag *Etag `json:"etag,omitempty"`
}

func (this *ModelBase) SetDefaults() {
	id, err := NewId()
	ft.PanicOnErr(err)
	safe.SetDefaultValue(&this.Id, *id)
	this.Etag = NewEtag()
}

func (this *ModelBase) ValidateRules(forEdit bool) []*val.FieldRules {
	return []*val.FieldRules{
		IdPtrValidateRule(&this.Id, forEdit),
		EtagPtrValidateRule(&this.Etag, forEdit),
	}
}

func (this *ModelBase) GetId() *Id {
	return this.Id
}

func (this *ModelBase) SetId(id Id) {
	this.Id = &id
}

func (this *ModelBase) GetEtag() *Etag {
	return this.Etag
}

func (this *ModelBase) SetEtag(etag Etag) {
	this.Etag = &etag
}

type AuditableBase struct {
	CreatedAt *time.Time `json:"createdAt,omitempty"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
}

func (this *AuditableBase) ValidateRules(forEdit bool) []*val.FieldRules {
	return []*val.FieldRules{}
}

func (this *AuditableBase) GetCreatedAt() *time.Time {
	return this.CreatedAt
}
func (this *AuditableBase) SetCreatedAt(createdAt time.Time) {
	this.CreatedAt = &createdAt
}
func (this *AuditableBase) GetUpdatedAt() *time.Time {
	return this.UpdatedAt
}
func (this *AuditableBase) SetUpdatedAt(updatedAt time.Time) {
	this.UpdatedAt = &updatedAt
}

type OrgBase struct {
	OrgId Id `json:"orgId,omitempty"`
}

func (this *OrgBase) ValidateRules(forEdit bool) []*val.FieldRules {
	return []*val.FieldRules{
		IdValidateRule(&this.OrgId, true),
	}
}

func (this *OrgBase) GetOrgId() Id {
	return this.OrgId
}
func (this *OrgBase) SetOrgId(orgId Id) {
	this.OrgId = orgId
}

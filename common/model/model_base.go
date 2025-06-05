package model

import (
	"time"

	util "github.com/sky-as-code/nikki-erp/common/util"
	val "github.com/sky-as-code/nikki-erp/common/validator"
)

type ModelBase struct {
	Id   *Id   `json:"id,omitempty"`
	Etag *Etag `json:"etag,omitempty"`
}

func (this *ModelBase) SetDefaults() error {
	id, err := NewId()
	if err != nil {
		return err
	}
	util.SetDefaultValue(this.Id, *id)
	this.Etag = NewEtag()
	return nil
}

func (this *ModelBase) ValidateRules(forEdit bool) []*val.FieldRules {
	return []*val.FieldRules{
		IdValidateRule(&this.Id, forEdit),
		EtagValidateRule(&this.Etag, forEdit),
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
	CreatedBy *Id        `json:"createdBy,omitempty"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
	UpdatedBy *Id        `json:"updatedBy,omitempty"`
}

func (this *AuditableBase) ValidateRules(forEdit bool) []*val.FieldRules {
	return []*val.FieldRules{
		IdValidateRule(&this.CreatedBy, !forEdit),
		IdValidateRule(&this.UpdatedBy, forEdit),
	}
}

func (this *AuditableBase) GetCreatedAt() *time.Time {
	return this.CreatedAt
}
func (this *AuditableBase) SetCreatedAt(createdAt time.Time) {
	this.CreatedAt = &createdAt
}
func (this *AuditableBase) GetCreatedBy() *Id {
	return this.CreatedBy
}
func (this *AuditableBase) SetCreatedBy(createdBy Id) {
	this.CreatedBy = &createdBy
}
func (this *AuditableBase) GetUpdatedAt() *time.Time {
	return this.UpdatedAt
}
func (this *AuditableBase) SetUpdatedAt(updatedAt time.Time) {
	this.UpdatedAt = &updatedAt
}
func (this *AuditableBase) GetUpdatedBy() *Id {
	return this.UpdatedBy
}
func (this *AuditableBase) SetUpdatedBy(updatedBy Id) {
	this.UpdatedBy = &updatedBy
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
func (this *OrgBase) IsOrgIdEmpty() *string {
	if this.OrgId == "" {
		return nil
	}
	orgStr := this.OrgId.String()
	return &orgStr
}

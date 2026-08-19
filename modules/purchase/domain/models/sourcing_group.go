package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	SourcingGroupSchemaName = "purchase_sourcing_group"

	SourcingGroupFieldId    = basemodel.FieldId
	SourcingGroupFieldEtag  = basemodel.FieldEtag
	SourcingGroupFieldOrgId = basemodel.FieldOrgId
)

//go:embed sourcing_group.json
var sourcingGroupSchemaJson string

func SourcingGroupSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(sourcingGroupSchemaJson)
}

type SourcingGroup struct {
	fields dmodel.DynamicFields
}

func NewSourcingGroup() *SourcingGroup {
	return &SourcingGroup{fields: make(dmodel.DynamicFields)}
}

func NewSourcingGroupFrom(src dmodel.DynamicFields) *SourcingGroup {
	return &SourcingGroup{fields: src}
}

func (this SourcingGroup) GetFieldData() dmodel.DynamicFields {
	return this.fields
}

func (this *SourcingGroup) SetFieldData(data dmodel.DynamicFields) {
	this.fields = data
}

func (this SourcingGroup) GetId() *model.Id {
	return this.fields.GetModelId(SourcingGroupFieldId)
}

func (this *SourcingGroup) SetId(v *model.Id) {
	this.fields.SetModelId(SourcingGroupFieldId, v)
}

func (this SourcingGroup) GetEtag() *model.Etag {
	return this.fields.GetEtag(SourcingGroupFieldEtag)
}

func (this *SourcingGroup) SetEtag(v *model.Etag) {
	this.fields.SetEtag(SourcingGroupFieldEtag, v)
}

func (this SourcingGroup) GetOrgId() *model.Id {
	return this.fields.GetModelId(SourcingGroupFieldOrgId)
}

func (this *SourcingGroup) SetOrgId(v *model.Id) {
	this.fields.SetModelId(SourcingGroupFieldOrgId, v)
}

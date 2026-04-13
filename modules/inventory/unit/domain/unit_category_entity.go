package domain

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	UnitCategorySchemaName = "inventory.unit_category"

	UnitCatFieldId    = basemodel.FieldId
	UnitCatFieldName  = "name"
	UnitCatFieldOrgId = "org_id"

	UnitCatEdgeUnits = "units"
)

func UnitCategorySchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(UnitCategorySchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Unit Category"}).
		TableName("invent_unit_categories").
		CompositeUnique(UnitCatFieldName, UnitCatFieldOrgId).
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			dmodel.DefineField().
				Name(UnitCatFieldName).
				Label(model.LangJson{model.LanguageCodeEnUs: "Name"}).
				DataType(dmodel.FieldDataTypeLangJson(1, model.MODEL_RULE_LONG_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(UnitCatFieldOrgId).
				Label(model.LangJson{model.LanguageCodeEnUs: "Organization"}).
				DataType(dmodel.FieldDataTypeUlid()).
				RequiredForCreate(),
		).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		EdgeFrom(
			dmodel.Edge(UnitCatEdgeUnits).
				Label(model.LangJson{model.LanguageCodeEnUs: "Units"}).
				Existing(UnitSchemaName, UnitEdgeCategory),
		)
}

type UnitCategory struct {
	fields dmodel.DynamicFields
}

func NewUnitCategory() *UnitCategory {
	return &UnitCategory{fields: make(dmodel.DynamicFields)}
}

func NewUnitCategoryFrom(src dmodel.DynamicFields) *UnitCategory {
	return &UnitCategory{fields: src}
}

func (this UnitCategory) GetFieldData() dmodel.DynamicFields {
	return this.fields
}

func (this *UnitCategory) SetFieldData(data dmodel.DynamicFields) {
	this.fields = data
}

func (this UnitCategory) GetId() *model.Id {
	return this.fields.GetModelId(basemodel.FieldId)
}

func (this *UnitCategory) SetId(v *model.Id) {
	this.fields.SetModelId(basemodel.FieldId, v)
}

func (this UnitCategory) GetName() *model.LangJson {
	v := this.fields.GetAny(UnitCatFieldName)
	if v == nil {
		return nil
	}
	lj := v.(model.LangJson)
	return &lj
}

func (this *UnitCategory) SetName(v *model.LangJson) {
	if v == nil {
		this.fields.SetAny(UnitCatFieldName, nil)
		return
	}
	this.fields.SetAny(UnitCatFieldName, *v)
}

func (this UnitCategory) GetOrgId() *model.Id {
	return this.fields.GetModelId(UnitCatFieldOrgId)
}

func (this *UnitCategory) SetOrgId(v *model.Id) {
	this.fields.SetModelId(UnitCatFieldOrgId, v)
}

func (this UnitCategory) GetEtag() *model.Etag {
	return this.fields.GetEtag(basemodel.FieldEtag)
}

func (this *UnitCategory) SetEtag(v *model.Etag) {
	this.fields.SetEtag(basemodel.FieldEtag, v)
}

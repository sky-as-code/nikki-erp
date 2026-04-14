package domain

import (
	"math"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

type UnitStatus string

const (
	UnitStatusDraft  = UnitStatus("draft")
	UnitStatusActive = UnitStatus("active")
)

func (this UnitStatus) String() string {
	return string(this)
}

func WrapUnitStatus(s string) *UnitStatus {
	st := UnitStatus(s)
	return &st
}

const (
	UnitSchemaName = "essential.unit"

	UnitFieldId         = basemodel.FieldId
	UnitFieldName       = "name"
	UnitFieldSymbol     = "symbol"
	UnitFieldBaseUnit   = "base_unit"
	UnitFieldMultiplier = "multiplier"
	UnitFieldStatus     = "status"
	UnitFieldCategoryId = "category_id"
	UnitFieldOrgId      = "org_id"

	UnitEdgeCategory = "unit_category"
)

func UnitSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(UnitSchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Unit"}).
		TableName("essential_units").
		CompositeUnique(UnitFieldSymbol, UnitFieldOrgId).
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			dmodel.DefineField().
				Name(UnitFieldName).
				Label(model.LangJson{model.LanguageCodeEnUs: "Name"}).
				DataType(dmodel.FieldDataTypeLangJson(1, model.MODEL_RULE_LONG_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(UnitFieldSymbol).
				Label(model.LangJson{model.LanguageCodeEnUs: "Symbol"}).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_SHORT_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(UnitFieldStatus).
				Label(model.LangJson{model.LanguageCodeEnUs: "Status"}).
				DataType(dmodel.FieldDataTypeEnumString([]string{
					string(UnitStatusDraft),
					string(UnitStatusActive),
				})).
				RequiredForCreate().
				Default(string(UnitStatusDraft)),
		).
		Field(
			dmodel.DefineField().
				Name(UnitFieldBaseUnit).
				Label(model.LangJson{model.LanguageCodeEnUs: "Base Unit"}).
				Description(model.LangJson{model.LanguageCodeEnUs: "Reference to the base unit for unit conversion"}).
				DataType(dmodel.FieldDataTypeUlid()),
		).
		Field(
			dmodel.DefineField().
				Name(UnitFieldMultiplier).
				Label(model.LangJson{model.LanguageCodeEnUs: "Multiplier"}).
				Description(model.LangJson{model.LanguageCodeEnUs: "Conversion multiplier relative to base unit"}).
				DataType(dmodel.FieldDataTypeInt64(1, math.MaxInt16)),
		).
		Field(
			dmodel.DefineField().
				Name(UnitFieldCategoryId).
				Label(model.LangJson{model.LanguageCodeEnUs: "Category"}).
				DataType(dmodel.FieldDataTypeUlid()),
		).
		Field(
			dmodel.DefineField().
				Name(UnitFieldOrgId).
				Label(model.LangJson{model.LanguageCodeEnUs: "Organization"}).
				DataType(dmodel.FieldDataTypeUlid()).
				RequiredForCreate(),
		).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		EdgeTo(
			dmodel.Edge(UnitEdgeCategory).
				Label(model.LangJson{model.LanguageCodeEnUs: "Unit Category"}).
				ManyToOne(UnitCategorySchemaName, dmodel.DynamicFields{
					UnitFieldCategoryId: basemodel.FieldId,
				}).
				OnDelete(dmodel.RelationCascadeSetNull),
		)
}

type Unit struct {
	fields dmodel.DynamicFields
}

func NewUnit() *Unit {
	return &Unit{fields: make(dmodel.DynamicFields)}
}

func NewUnitFrom(src dmodel.DynamicFields) *Unit {
	return &Unit{fields: src}
}

func (this Unit) GetFieldData() dmodel.DynamicFields {
	return this.fields
}

func (this *Unit) SetFieldData(data dmodel.DynamicFields) {
	this.fields = data
}

func (this Unit) GetId() *model.Id {
	return this.fields.GetModelId(basemodel.FieldId)
}

func (this *Unit) SetId(v *model.Id) {
	this.fields.SetModelId(basemodel.FieldId, v)
}

func (this Unit) GetName() *model.LangJson {
	v := this.fields.GetAny(UnitFieldName)
	if v == nil {
		return nil
	}
	lj := v.(model.LangJson)
	return &lj
}

func (this *Unit) SetName(v *model.LangJson) {
	if v == nil {
		this.fields.SetAny(UnitFieldName, nil)
		return
	}
	this.fields.SetAny(UnitFieldName, *v)
}

func (this Unit) GetSymbol() *string {
	return this.fields.GetString(UnitFieldSymbol)
}

func (this *Unit) SetSymbol(v *string) {
	this.fields.SetString(UnitFieldSymbol, v)
}

func (this Unit) GetStatus() *UnitStatus {
	s := this.fields.GetString(UnitFieldStatus)
	if s == nil {
		return nil
	}
	st := UnitStatus(*s)
	return &st
}

func (this Unit) GetBaseUnit() *model.Id {
	return this.fields.GetModelId(UnitFieldBaseUnit)
}

func (this *Unit) SetBaseUnit(v *model.Id) {
	this.fields.SetModelId(UnitFieldBaseUnit, v)
}

func (this Unit) GetMultiplier() *int64 {
	return this.fields.GetInt64(UnitFieldMultiplier)
}

func (this Unit) GetCategoryId() *model.Id {
	return this.fields.GetModelId(UnitFieldCategoryId)
}

func (this Unit) GetOrgId() *model.Id {
	return this.fields.GetModelId(UnitFieldOrgId)
}

func (this *Unit) SetOrgId(v *model.Id) {
	this.fields.SetModelId(UnitFieldOrgId, v)
}

func (this *Unit) SetStatus(v *UnitStatus) {
	if v == nil {
		this.fields.SetString(UnitFieldStatus, nil)
		return
	}
	s := string(*v)
	this.fields.SetString(UnitFieldStatus, &s)
}

func (this *Unit) SetMultiplier(v *int64) {
	this.fields.SetInt64(UnitFieldMultiplier, v)
}

func (this *Unit) SetCategoryId(v *model.Id) {
	this.fields.SetModelId(UnitFieldCategoryId, v)
}

func (this Unit) GetEtag() *model.Etag {
	return this.fields.GetEtag(basemodel.FieldEtag)
}

func (this *Unit) SetEtag(v *model.Etag) {
	this.fields.SetEtag(basemodel.FieldEtag, v)
}

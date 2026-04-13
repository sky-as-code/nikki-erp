package domain

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/json"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

type ProductStatus string

const (
	ProductStatusActive   = ProductStatus("active")
	ProductStatusArchived = ProductStatus("archived")
)

func (this ProductStatus) String() string {
	return string(this)
}

func WrapProductStatus(s string) *ProductStatus {
	st := ProductStatus(s)
	return &st
}

const (
	ProductSchemaName = "inventory.product"

	ProdFieldId               = basemodel.FieldId
	ProdFieldName             = "name"
	ProdFieldDescription      = "description"
	ProdFieldStatus           = "status"
	ProdFieldThumbnailUrl     = "thumbnail_url"
	ProdFieldUnitId           = "unit_id"
	ProdFieldDefaultVariantId = "default_variant_id"
	ProdFieldTagIds           = "tag_ids"
	ProdFieldOrgId            = "org_id"

	ProdEdgeCategories      = "categories"
	ProdEdgeVariants        = "variants"
	ProdEdgeAttributeGroups = "attribute_groups"
	ProdEdgeAttributes      = "attributes"
)

func ProductSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(ProductSchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Product"}).
		TableName("invent_products").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			dmodel.DefineField().
				Name(ProdFieldName).
				Label(model.LangJson{model.LanguageCodeEnUs: "Name"}).
				DataType(dmodel.FieldDataTypeLangJson(1, model.MODEL_RULE_LONG_NAME_LENGTH)).
				Unique().
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(ProdFieldDescription).
				Label(model.LangJson{model.LanguageCodeEnUs: "Description"}).
				DataType(dmodel.FieldDataTypeLangJson(0, model.MODEL_RULE_DESC_LENGTH)),
		).
		Field(
			dmodel.DefineField().
				Name(ProdFieldOrgId).
				Label(model.LangJson{model.LanguageCodeEnUs: "Organization"}).
				DataType(dmodel.FieldDataTypeUlid()).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(ProdFieldStatus).
				Label(model.LangJson{model.LanguageCodeEnUs: "Status"}).
				DataType(dmodel.FieldDataTypeEnumString([]string{
					string(ProductStatusActive),
					string(ProductStatusArchived),
				})).
				Default(string(ProductStatusArchived)),
		).
		Field(
			dmodel.DefineField().
				Name(ProdFieldThumbnailUrl).
				Label(model.LangJson{model.LanguageCodeEnUs: "Thumbnail URL"}).
				DataType(dmodel.FieldDataTypeUrl()),
		).
		Field(
			dmodel.DefineField().
				Name(ProdFieldUnitId).
				Label(model.LangJson{model.LanguageCodeEnUs: "Unit"}).
				DataType(dmodel.FieldDataTypeUlid()),
		).
		Field(
			dmodel.DefineField().
				Name(ProdFieldDefaultVariantId).
				Label(model.LangJson{model.LanguageCodeEnUs: "Default Variant"}).
				DataType(dmodel.FieldDataTypeUlid()),
		).
		Field(
			dmodel.DefineField().
				Name(ProdFieldTagIds).
				Label(model.LangJson{model.LanguageCodeEnUs: "Tag IDs"}).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_LONG_NAME_LENGTH)),
		).
		Extend(basemodel.ArchivableModelSchemaBuilder()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		EdgeTo(
			dmodel.Edge(ProdEdgeCategories).
				Label(model.LangJson{model.LanguageCodeEnUs: "Categories"}).
				ManyToMany(ProductCategorySchemaName, ProdCatRelSchemaName, "product").
				OnDelete(dmodel.RelationCascadeCascade),
		).
		EdgeFrom(
			dmodel.Edge(ProdEdgeVariants).
				Label(model.LangJson{model.LanguageCodeEnUs: "Variants"}).
				Existing(VariantSchemaName, VarEdgeProduct),
		).
		EdgeFrom(
			dmodel.Edge(ProdEdgeAttributeGroups).
				Label(model.LangJson{model.LanguageCodeEnUs: "Attribute Groups"}).
				Existing(AttributeGroupSchemaName, AttrGrpEdgeProduct),
		).
		EdgeFrom(
			dmodel.Edge(ProdEdgeAttributes).
				Label(model.LangJson{model.LanguageCodeEnUs: "Attributes"}).
				Existing(AttributeSchemaName, AttrEdgeProduct),
		)
}

type Product struct {
	fields dmodel.DynamicFields
}

func NewProduct() *Product {
	return &Product{fields: make(dmodel.DynamicFields)}
}

func NewProductFrom(src dmodel.DynamicFields) *Product {
	return &Product{fields: src}
}

func (this Product) GetFieldData() dmodel.DynamicFields {
	return this.fields
}

func (this *Product) SetFieldData(data dmodel.DynamicFields) {
	this.fields = data
}

func (this Product) GetId() *model.Id {
	return this.fields.GetModelId(basemodel.FieldId)
}

func (this *Product) SetId(v *model.Id) {
	this.fields.SetModelId(basemodel.FieldId, v)
}

func (this Product) IsArchived() bool {
	val := this.fields.GetBool(basemodel.FieldIsArchived)
	if val == nil {
		return false
	}
	return *val
}

func (this *Product) SetIsArchived(v *bool) {
	this.fields.SetBool(basemodel.FieldIsArchived, v)
}

func (this Product) GetName() *model.LangJson {
	v := this.fields.GetAny(ProdFieldName)
	if v == nil {
		return nil
	}
	if strVal, ok := v.(string); ok {
		var langJson model.LangJson
		if err := json.UnmarshalStr(strVal, &langJson); err == nil {
			return &langJson
		}
	}
	return nil
}

func (this *Product) SetName(v *model.LangJson) {
	if v == nil {
		this.fields.SetAny(ProdFieldName, nil)
		return
	}
	this.fields.SetAny(ProdFieldName, *v)
}

func (this Product) GetDescription() *model.LangJson {
	val := this.fields.GetAny(ProdFieldDescription)
	if val == nil {
		return nil
	}

	if strVal, ok := val.(string); ok {
		var langJson model.LangJson
		if err := json.UnmarshalStr(strVal, &langJson); err == nil {
			return &langJson
		}
	}
	return nil
}

func (this *Product) SetDescription(v *model.LangJson) {
	if v == nil {
		this.fields.SetAny(ProdFieldDescription, nil)
		return
	}
	this.fields.SetAny(ProdFieldDescription, *v)
}

func (this Product) GetStatus() *ProductStatus {
	s := this.fields.GetString(ProdFieldStatus)
	if s == nil {
		return nil
	}
	st := ProductStatus(*s)
	return &st
}

func (this *Product) SetStatus(v *ProductStatus) {
	if v == nil {
		this.fields.SetString(ProdFieldStatus, nil)
		return
	}
	s := string(*v)
	this.fields.SetString(ProdFieldStatus, &s)
}

func (this Product) GetThumbnailUrl() *string {
	return this.fields.GetString(ProdFieldThumbnailUrl)
}

func (this *Product) SetThumbnailUrl(v *string) {
	this.fields.SetString(ProdFieldThumbnailUrl, v)
}

func (this Product) GetUnitId() *model.Id {
	return this.fields.GetModelId(ProdFieldUnitId)
}

func (this *Product) SetUnitId(v *model.Id) {
	this.fields.SetModelId(ProdFieldUnitId, v)
}

func (this Product) GetDefaultVariantId() *model.Id {
	return this.fields.GetModelId(ProdFieldDefaultVariantId)
}

func (this *Product) SetDefaultVariantId(v *model.Id) {
	this.fields.SetModelId(ProdFieldDefaultVariantId, v)
}

func (this Product) GetTagIds() *string {
	return this.fields.GetString(ProdFieldTagIds)
}

func (this *Product) SetTagIds(v *string) {
	this.fields.SetString(ProdFieldTagIds, v)
}

func (this Product) GetOrgId() *model.Id {
	return this.fields.GetModelId(ProdFieldOrgId)
}

func (this *Product) SetOrgId(v *model.Id) {
	this.fields.SetModelId(ProdFieldOrgId, v)
}

func (this Product) GetEtag() *model.Etag {
	return this.fields.GetEtag(basemodel.FieldEtag)
}

func (this *Product) SetEtag(v *model.Etag) {
	this.fields.SetEtag(basemodel.FieldEtag, v)
}

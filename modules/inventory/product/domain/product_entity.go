package domain

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/json"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	ProductSchemaName = "inventory.product"

	ProdFieldId               = basemodel.FieldId
	ProdFieldName             = "name"
	ProdFieldDescription      = "description"
	ProdFieldThumbnailUrl     = "thumbnail_url"
	ProdFieldUnitId           = "unit_id"
	ProdFieldDefaultVariantId = "default_variant_id"
	ProdFieldTagIds           = "tag_ids"

	ProdEdgeCategories      = "categories"
	ProdEdgeVariants        = "variants"
	ProdEdgeAttributeGroups = "attribute_groups"
	ProdEdgeAttributes      = "attributes"
)

func ProductSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(ProductSchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Product"}).
		TableName("inventory_products").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Extend(basemodel.OrgIdModelSchemaBuilder()).
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
				Name(ProdFieldThumbnailUrl).
				Label(model.LangJson{model.LanguageCodeEnUs: "Thumbnail URL"}).
				DataType(dmodel.FieldDataTypeUrl()),
		).
		Field(
			basemodel.DefineFieldId(ProdFieldUnitId).
				Label(model.LangJson{model.LanguageCodeEnUs: "Unit"}),
		).
		Field(
			basemodel.DefineFieldId(ProdFieldDefaultVariantId).
				Label(model.LangJson{model.LanguageCodeEnUs: "Default Variant"}),
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
	basemodel.DynamicModelBase
}

func NewProduct() *Product {
	return &Product{basemodel.NewDynamicModel()}
}

func NewProductFrom(src dmodel.DynamicFields) *Product {
	return &Product{basemodel.NewDynamicModel(src)}
}

func (this Product) GetName() *model.LangJson {
	v := this.GetFieldData().GetAny(ProdFieldName)
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
		this.GetFieldData().SetAny(ProdFieldName, nil)
		return
	}
	this.GetFieldData().SetAny(ProdFieldName, *v)
}

func (this Product) GetDescription() *model.LangJson {
	val := this.GetFieldData().GetAny(ProdFieldDescription)
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
		this.GetFieldData().SetAny(ProdFieldDescription, nil)
		return
	}
	this.GetFieldData().SetAny(ProdFieldDescription, *v)
}

func (this Product) GetThumbnailUrl() *string {
	return this.GetFieldData().GetString(ProdFieldThumbnailUrl)
}

func (this *Product) SetThumbnailUrl(v *string) {
	this.GetFieldData().SetString(ProdFieldThumbnailUrl, v)
}

func (this Product) GetUnitId() *model.Id {
	return this.GetFieldData().GetModelId(ProdFieldUnitId)
}

func (this *Product) SetUnitId(v *model.Id) {
	this.GetFieldData().SetModelId(ProdFieldUnitId, v)
}

func (this Product) GetDefaultVariantId() *model.Id {
	return this.GetFieldData().GetModelId(ProdFieldDefaultVariantId)
}

func (this *Product) SetDefaultVariantId(v *model.Id) {
	this.GetFieldData().SetModelId(ProdFieldDefaultVariantId, v)
}

func (this Product) GetTagIds() *string {
	return this.GetFieldData().GetString(ProdFieldTagIds)
}

func (this *Product) SetTagIds(v *string) {
	this.GetFieldData().SetString(ProdFieldTagIds, v)
}

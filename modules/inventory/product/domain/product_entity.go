package domain

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	ProductResourceCode = "inventory_product"
	ProductAuthScope    = "org"

	ProductActionCreate      = "create"
	ProductActionDelete      = "delete"
	ProductActionUpdate      = "update"
	ProductActionView        = "view"
	ProductActionSetArchived = "set_archived"
)

const (
	ProductSchemaName = "inventory.product"

	ProdFieldId                = basemodel.FieldId
	ProdFieldName              = "name"
	ProdFieldDescription       = "description"
	ProdFieldThumbnailMediaRef = "media_ref"
	ProdFieldThumbnail         = "thumbnail"
	ProdFieldUnitId            = "unit_id"
	ProdFieldDefaultVariantId  = "default_variant_id"
	ProdFieldTagIds            = "tag_ids"

	ProdEdgeCategories      = "categories"
	ProdEdgeVariants        = "variants"
	ProdEdgeAttributeGroups = "attribute_groups"
	ProdEdgeAttributes      = "attributes"
	ProdEdgeAttributeValues = "attribute_values"
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
				Name(ProdFieldThumbnailMediaRef).
				Label(model.LangJson{model.LanguageCodeEnUs: "Thumbnail Media"}).
				DataType(dmodel.FieldDataTypeUlid()),
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
		).
		EdgeFrom(
			dmodel.Edge(ProdEdgeAttributeValues).
				Label(model.LangJson{model.LanguageCodeEnUs: "Attribute Values"}).
				Existing(AttributeValueSchemaName, AttrValEdgeProduct),
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
	return this.GetFieldData().GetLangJson(ProdFieldName)
}

func (this *Product) SetName(v *model.LangJson) {
	this.GetFieldData().SetLangJson(ProdFieldName, v)
}

func (this Product) GetDescription() *model.LangJson {
	return this.GetFieldData().GetLangJson(ProdFieldDescription)
}

func (this *Product) SetDescription(v *model.LangJson) {
	this.GetFieldData().SetLangJson(ProdFieldDescription, v)
}

func (this Product) GetUnitId() *model.Id {
	return this.GetFieldData().GetModelId(ProdFieldUnitId)
}

func (this Product) GetThumbnailMediaRef() *string {
	return this.GetFieldData().GetString(ProdFieldThumbnailMediaRef)
}

func (this *Product) SetThumbnailMediaRef(v *string) {
	this.GetFieldData().SetString(ProdFieldThumbnailMediaRef, v)
}

func (this Product) GetThumbnail() dmodel.DynamicFields {
	raw, ok := this.GetFieldData().GetAny(ProdFieldThumbnail).(dmodel.DynamicFields)
	if !ok {
		return nil
	}
	return raw
}

func (this *Product) SetThumbnail(v dmodel.DynamicFields) {
	this.GetFieldData().SetAny(ProdFieldThumbnail, v)
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

func (this *Product) GetCategories() []*ProductCategory {
	data := this.GetFieldData().GetAny(ProdEdgeCategories)
	raws, ok := data.([]dmodel.DynamicFields)
	if !ok {
		return nil
	}

	res := []*ProductCategory{}
	for _, raw := range raws {
		cat := NewProductCategoryFrom(raw)
		res = append(res, cat)
	}

	return res
}

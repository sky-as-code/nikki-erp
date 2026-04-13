package domain

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	ProductCategorySchemaName = "inventory.product_category"

	ProductCategoryFieldId = basemodel.FieldId
	ProdCatFieldName       = "name"
	ProdCatFieldOrgId      = "org_id"

	ProdCatEdgeProducts = "products"

	ProdCatRelSchemaName             = "inventory.product_category_rel"
	ProdCatRelFieldProductId         = "product_id"
	ProdCatRelFieldProductCategoryId = "product_category_id"
)

func ProductCategoryRelSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(ProdCatRelSchemaName).
		TableName("invent_product_category_rel").
		ShouldBuildDb().
		Field(
			dmodel.DefineField().
				Name(ProdCatRelFieldProductId).
				DataType(dmodel.FieldDataTypeUlid()).
				PrimaryKey(),
		).
		Field(
			dmodel.DefineField().
				Name(ProdCatRelFieldProductCategoryId).
				DataType(dmodel.FieldDataTypeUlid()).
				PrimaryKey(),
		)
}

func ProductCategorySchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(ProductCategorySchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Product Category"}).
		TableName("invent_product_categories").
		CompositeUnique(ProdCatFieldName, ProdCatFieldOrgId).
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			dmodel.DefineField().
				Name(ProdCatFieldName).
				Label(model.LangJson{model.LanguageCodeEnUs: "Name"}).
				DataType(dmodel.FieldDataTypeLangJson(1, model.MODEL_RULE_LONG_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(ProdCatFieldOrgId).
				Label(model.LangJson{model.LanguageCodeEnUs: "Organization"}).
				DataType(dmodel.FieldDataTypeUlid()).
				RequiredForCreate(),
		).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		EdgeTo(
			dmodel.Edge(ProdCatEdgeProducts).
				Label(model.LangJson{model.LanguageCodeEnUs: "Products"}).
				ManyToMany(ProductSchemaName, ProdCatRelSchemaName, "product_category").
				OnDelete(dmodel.RelationCascadeCascade),
		)
}

type ProductCategory struct {
	fields dmodel.DynamicFields
}

func NewProductCategory() *ProductCategory {
	return &ProductCategory{fields: make(dmodel.DynamicFields)}
}

func NewProductCategoryFrom(src dmodel.DynamicFields) *ProductCategory {
	return &ProductCategory{fields: src}
}

func (this ProductCategory) GetFieldData() dmodel.DynamicFields {
	return this.fields
}

func (this *ProductCategory) SetFieldData(data dmodel.DynamicFields) {
	this.fields = data
}

func (this ProductCategory) GetId() *model.Id {
	return this.fields.GetModelId(basemodel.FieldId)
}

func (this *ProductCategory) SetId(v *model.Id) {
	this.fields.SetModelId(basemodel.FieldId, v)
}

func (this ProductCategory) GetName() *model.LangJson {
	v := this.fields.GetAny(ProdCatFieldName)
	if v == nil {
		return nil
	}
	lj := v.(model.LangJson)
	return &lj
}

func (this *ProductCategory) SetName(v *model.LangJson) {
	if v == nil {
		this.fields.SetAny(ProdCatFieldName, nil)
		return
	}
	this.fields.SetAny(ProdCatFieldName, *v)
}

func (this ProductCategory) GetOrgId() *model.Id {
	return this.fields.GetModelId(ProdCatFieldOrgId)
}

func (this *ProductCategory) SetOrgId(v *model.Id) {
	this.fields.SetModelId(ProdCatFieldOrgId, v)
}

func (this ProductCategory) GetEtag() *model.Etag {
	return this.fields.GetEtag(basemodel.FieldEtag)
}

func (this *ProductCategory) SetEtag(v *model.Etag) {
	this.fields.SetEtag(basemodel.FieldEtag, v)
}

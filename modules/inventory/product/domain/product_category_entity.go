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

	ProdCatEdgeProducts = "products"

	ProdCatRelSchemaName             = "inventory.product_category_rel"
	ProdCatRelFieldProductId         = "product_id"
	ProdCatRelFieldProductCategoryId = "product_category_id"
)

func ProductCategoryRelSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(ProdCatRelSchemaName).
		TableName("inventory_product_category_rel").
		ShouldBuildDb().
		ExtendBase().
		Field(
			basemodel.DefineFieldId(ProdCatRelFieldProductId).
				PrimaryKey(),
		).
		Field(
			basemodel.DefineFieldId(ProdCatRelFieldProductCategoryId).
				PrimaryKey(),
		)
}

func ProductCategorySchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(ProductCategorySchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Product Category"}).
		TableName("inventory_product_categories").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Extend(basemodel.OrgIdModelSchemaBuilder()).
		CompositeUnique(ProdCatFieldName, basemodel.FieldOrgId).
		Field(
			dmodel.DefineField().
				Name(ProdCatFieldName).
				Label(model.LangJson{model.LanguageCodeEnUs: "Name"}).
				DataType(dmodel.FieldDataTypeLangJson(1, model.MODEL_RULE_LONG_NAME_LENGTH)).
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
	basemodel.DynamicModelBase
}

func NewProductCategory() *ProductCategory {
	return &ProductCategory{basemodel.NewDynamicModel()}
}

func NewProductCategoryFrom(src dmodel.DynamicFields) *ProductCategory {
	return &ProductCategory{basemodel.NewDynamicModel(src)}
}

func (this ProductCategory) GetName() *model.LangJson {
	v := this.GetFieldData().GetAny(ProdCatFieldName)
	if v == nil {
		return nil
	}
	lj := v.(model.LangJson)
	return &lj
}

func (this *ProductCategory) SetName(v *model.LangJson) {
	if v == nil {
		this.GetFieldData().SetAny(ProdCatFieldName, nil)
		return
	}
	this.GetFieldData().SetAny(ProdCatFieldName, *v)
}

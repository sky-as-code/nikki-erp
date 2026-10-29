package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	ProductCategorySchemaName = "inventory_product_category"

	ProductCategoryFieldId               = basemodel.FieldId
	ProductCategoryFieldName             = "name"
	ProductCategoryFieldParentCategoryId = "parent_category_id"
	ProductCategoryFieldIsRoot           = "is_root"
	ProductCategoryFieldCanonicalName    = "canonical_name"
	ProductCategoryFieldDisplayOrder     = "display_order"
	ProductCategoryFieldDescription      = "description"
	ProductCategoryFieldOrgId            = "org_id"

	ProductCategoryEdgeParentCategory = "parent_category"

	// ProductCategoryCanonicalSeparator joins the segments of canonical_name.
	ProductCategoryCanonicalSeparator = " / "
)

// ProductCategoryRootName is the name every org's root category carries.
func ProductCategoryRootName() model.LangJson {
	return model.LangJson{model.LanguageCodeEnUs: "Default", model.LanguageCodeViVn: "Mặc định"}
}

//go:embed product_category.json
var productCategorySchemaJson string

func ProductCategorySchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(productCategorySchemaJson)
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
	return this.GetFieldData().GetLangJson(ProductCategoryFieldName)
}

func (this *ProductCategory) SetName(v *model.LangJson) {
	this.GetFieldData().SetLangJson(ProductCategoryFieldName, v)
}

// GetParentCategoryId is the category's own id for the org's root.
func (this ProductCategory) GetParentCategoryId() *model.Id {
	return this.GetFieldData().GetModelId(ProductCategoryFieldParentCategoryId)
}

func (this *ProductCategory) SetParentCategoryId(v *model.Id) {
	this.GetFieldData().SetModelId(ProductCategoryFieldParentCategoryId, v)
}

func (this ProductCategory) GetIsRoot() *bool {
	return this.GetFieldData().GetBool(ProductCategoryFieldIsRoot)
}

func (this ProductCategory) IsRoot() bool {
	isRoot := this.GetIsRoot()
	return isRoot != nil && *isRoot
}

func (this ProductCategory) GetCanonicalName() *model.LangJson {
	return this.GetFieldData().GetLangJson(ProductCategoryFieldCanonicalName)
}

func (this *ProductCategory) SetCanonicalName(v *model.LangJson) {
	this.GetFieldData().SetLangJson(ProductCategoryFieldCanonicalName, v)
}

func (this ProductCategory) GetDisplayOrder() *int32 {
	return this.GetFieldData().GetInt32(ProductCategoryFieldDisplayOrder)
}

func (this *ProductCategory) SetDisplayOrder(v *int32) {
	this.GetFieldData().SetInt32(ProductCategoryFieldDisplayOrder, v)
}

func (this ProductCategory) GetDescription() *string {
	return this.GetFieldData().GetString(ProductCategoryFieldDescription)
}

func (this *ProductCategory) SetDescription(v *string) {
	this.GetFieldData().SetString(ProductCategoryFieldDescription, v)
}

func (this ProductCategory) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(ProductCategoryFieldOrgId)
}

func (this *ProductCategory) SetOrgId(v *model.Id) {
	this.GetFieldData().SetModelId(ProductCategoryFieldOrgId, v)
}

package models

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// ProductTemplateStatus is the business lifecycle of a product line. It is deliberately separate
// from is_archived, which is system visibility: a discontinued product may stay unarchived so it
// still appears in discontinued-product listings. See BR §2.3 and AC-PROD-018.
type ProductTemplateStatus string

const (
	ProductTemplateStatusDraft        = ProductTemplateStatus("draft")
	ProductTemplateStatusActive       = ProductTemplateStatus("active")
	ProductTemplateStatusDiscontinued = ProductTemplateStatus("discontinued")
)

func (this ProductTemplateStatus) String() string {
	return string(this)
}

func WrapProductTemplateStatus(s string) *ProductTemplateStatus {
	st := ProductTemplateStatus(s)
	return &st
}

const (
	ProductTemplateSchemaName = "inventory_product_template"

	ProductTemplateFieldId                  = basemodel.FieldId
	ProductTemplateFieldName                = "name"
	ProductTemplateFieldShortName           = "short_name"
	ProductTemplateFieldProductTypeId       = "product_type_id"
	ProductTemplateFieldCategoryId          = "category_id"
	ProductTemplateFieldBrandId             = "brand_id"
	ProductTemplateFieldSaleOk              = "sale_ok"
	ProductTemplateFieldPurchaseOk          = "purchase_ok"
	ProductTemplateFieldDescription         = "description"
	ProductTemplateFieldSalesDescription    = "sales_description"
	ProductTemplateFieldPurchaseDescription = "purchase_description"
	ProductTemplateFieldDefaultImageId      = "default_image_id"
	ProductTemplateFieldBaseSalesPrice      = "base_sales_price"
	ProductTemplateFieldDefaultWeight       = "default_weight"
	ProductTemplateFieldDefaultLength       = "default_length"
	ProductTemplateFieldDefaultWidth        = "default_width"
	ProductTemplateFieldDefaultHeight       = "default_height"
	ProductTemplateFieldStatus              = "status"
	ProductTemplateFieldOrgId               = "org_id"

	// The cost read model. A template has no cost of its own — cost belongs to the concrete
	// variant — so these expose the RANGE across its variants and nothing more. Never treat
	// either as the product's cost (BR-PRICE-VARIANT-014).
	ProductTemplateFieldMinVariantCost = "min_variant_cost"
	ProductTemplateFieldMaxVariantCost = "max_variant_cost"

	ProductTemplateEdgeProductType = "product_type"
	ProductTemplateEdgeCategory    = "category"
	ProductTemplateEdgeBrand       = "brand"
	ProductTemplateEdgeVariants    = "variants"
)

//go:embed product_template.json
var productTemplateSchemaJson string

func ProductTemplateSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(productTemplateSchemaJson)
}

type ProductTemplate struct {
	basemodel.DynamicModelBase
}

func NewProductTemplate() *ProductTemplate {
	return &ProductTemplate{basemodel.NewDynamicModel()}
}

func NewProductTemplateFrom(src dmodel.DynamicFields) *ProductTemplate {
	return &ProductTemplate{basemodel.NewDynamicModel(src)}
}

func (this ProductTemplate) GetName() *model.LangJson {
	return this.GetFieldData().GetLangJson(ProductTemplateFieldName)
}

func (this *ProductTemplate) SetName(v *model.LangJson) {
	this.GetFieldData().SetLangJson(ProductTemplateFieldName, v)
}

func (this ProductTemplate) GetShortName() *model.LangJson {
	return this.GetFieldData().GetLangJson(ProductTemplateFieldShortName)
}

func (this *ProductTemplate) SetShortName(v *model.LangJson) {
	this.GetFieldData().SetLangJson(ProductTemplateFieldShortName, v)
}

func (this ProductTemplate) GetProductTypeId() *model.Id {
	return this.GetFieldData().GetModelId(ProductTemplateFieldProductTypeId)
}

func (this *ProductTemplate) SetProductTypeId(v *model.Id) {
	this.GetFieldData().SetModelId(ProductTemplateFieldProductTypeId, v)
}

func (this ProductTemplate) GetCategoryId() *model.Id {
	return this.GetFieldData().GetModelId(ProductTemplateFieldCategoryId)
}

func (this *ProductTemplate) SetCategoryId(v *model.Id) {
	this.GetFieldData().SetModelId(ProductTemplateFieldCategoryId, v)
}

func (this ProductTemplate) GetBrandId() *model.Id {
	return this.GetFieldData().GetModelId(ProductTemplateFieldBrandId)
}

func (this *ProductTemplate) SetBrandId(v *model.Id) {
	this.GetFieldData().SetModelId(ProductTemplateFieldBrandId, v)
}

func (this ProductTemplate) GetSaleOk() *bool {
	return this.GetFieldData().GetBool(ProductTemplateFieldSaleOk)
}

func (this *ProductTemplate) SetSaleOk(v *bool) {
	this.GetFieldData().SetBool(ProductTemplateFieldSaleOk, v)
}

func (this ProductTemplate) GetPurchaseOk() *bool {
	return this.GetFieldData().GetBool(ProductTemplateFieldPurchaseOk)
}

func (this *ProductTemplate) SetPurchaseOk(v *bool) {
	this.GetFieldData().SetBool(ProductTemplateFieldPurchaseOk, v)
}

func (this ProductTemplate) GetDescription() *model.LangJson {
	return this.GetFieldData().GetLangJson(ProductTemplateFieldDescription)
}

func (this *ProductTemplate) SetDescription(v *model.LangJson) {
	this.GetFieldData().SetLangJson(ProductTemplateFieldDescription, v)
}

func (this ProductTemplate) GetSalesDescription() *model.LangJson {
	return this.GetFieldData().GetLangJson(ProductTemplateFieldSalesDescription)
}

func (this *ProductTemplate) SetSalesDescription(v *model.LangJson) {
	this.GetFieldData().SetLangJson(ProductTemplateFieldSalesDescription, v)
}

func (this ProductTemplate) GetPurchaseDescription() *model.LangJson {
	return this.GetFieldData().GetLangJson(ProductTemplateFieldPurchaseDescription)
}

func (this *ProductTemplate) SetPurchaseDescription(v *model.LangJson) {
	this.GetFieldData().SetLangJson(ProductTemplateFieldPurchaseDescription, v)
}

func (this ProductTemplate) GetDefaultImageId() *model.Id {
	return this.GetFieldData().GetModelId(ProductTemplateFieldDefaultImageId)
}

func (this *ProductTemplate) SetDefaultImageId(v *model.Id) {
	this.GetFieldData().SetModelId(ProductTemplateFieldDefaultImageId, v)
}

func (this ProductTemplate) GetBaseSalesPrice() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(ProductTemplateFieldBaseSalesPrice)
}

func (this *ProductTemplate) SetBaseSalesPrice(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(ProductTemplateFieldBaseSalesPrice, v)
}

func (this ProductTemplate) GetDefaultWeight() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(ProductTemplateFieldDefaultWeight)
}

func (this *ProductTemplate) SetDefaultWeight(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(ProductTemplateFieldDefaultWeight, v)
}

func (this ProductTemplate) GetDefaultLength() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(ProductTemplateFieldDefaultLength)
}

func (this *ProductTemplate) SetDefaultLength(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(ProductTemplateFieldDefaultLength, v)
}

func (this ProductTemplate) GetDefaultWidth() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(ProductTemplateFieldDefaultWidth)
}

func (this *ProductTemplate) SetDefaultWidth(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(ProductTemplateFieldDefaultWidth, v)
}

func (this ProductTemplate) GetDefaultHeight() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(ProductTemplateFieldDefaultHeight)
}

func (this *ProductTemplate) SetDefaultHeight(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(ProductTemplateFieldDefaultHeight, v)
}

func (this ProductTemplate) GetStatus() *ProductTemplateStatus {
	s := this.GetFieldData().GetString(ProductTemplateFieldStatus)
	if s == nil {
		return nil
	}
	return WrapProductTemplateStatus(*s)
}

func (this *ProductTemplate) SetStatus(v *ProductTemplateStatus) {
	if v == nil {
		this.GetFieldData().SetString(ProductTemplateFieldStatus, nil)
		return
	}
	s := string(*v)
	this.GetFieldData().SetString(ProductTemplateFieldStatus, &s)
}

func (this ProductTemplate) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(ProductTemplateFieldOrgId)
}

func (this *ProductTemplate) SetOrgId(v *model.Id) {
	this.GetFieldData().SetModelId(ProductTemplateFieldOrgId, v)
}

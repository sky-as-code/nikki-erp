package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// Product Type determines how the system processes a product, as opposed to how the business
// classifies it. Taxonomy such as Raw Material or Finished Goods belongs on Product Category,
// not here. See BR §6.3.1.
const (
	ProductTypeCodeGoods   = "GOODS"
	ProductTypeCodeService = "SERVICE"
	ProductTypeCodeCombo   = "COMBO"
)

const (
	ProductTypeSchemaName = "inventory_product_type"

	ProductTypeFieldId                    = basemodel.FieldId
	ProductTypeFieldCode                  = "code"
	ProductTypeFieldName                  = "name"
	ProductTypeFieldDescription           = "description"
	ProductTypeFieldSupportsStock         = "supports_stock"
	ProductTypeFieldSupportsSale          = "supports_sale"
	ProductTypeFieldSupportsPurchase      = "supports_purchase"
	ProductTypeFieldSupportsManufacturing = "supports_manufacturing"
)

//go:embed product_type.json
var productTypeSchemaJson string

func ProductTypeSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(productTypeSchemaJson)
}

type ProductType struct {
	basemodel.DynamicModelBase
}

func NewProductType() *ProductType {
	return &ProductType{basemodel.NewDynamicModel()}
}

func NewProductTypeFrom(src dmodel.DynamicFields) *ProductType {
	return &ProductType{basemodel.NewDynamicModel(src)}
}

func (this ProductType) GetCode() *string {
	return this.GetFieldData().GetString(ProductTypeFieldCode)
}

func (this *ProductType) SetCode(v *string) {
	this.GetFieldData().SetString(ProductTypeFieldCode, v)
}

func (this ProductType) GetName() *model.LangJson {
	return this.GetFieldData().GetLangJson(ProductTypeFieldName)
}

func (this *ProductType) SetName(v *model.LangJson) {
	this.GetFieldData().SetLangJson(ProductTypeFieldName, v)
}

func (this ProductType) GetDescription() *model.LangJson {
	return this.GetFieldData().GetLangJson(ProductTypeFieldDescription)
}

func (this *ProductType) SetDescription(v *model.LangJson) {
	this.GetFieldData().SetLangJson(ProductTypeFieldDescription, v)
}

// GetSupportsStock reports whether products of this type take part in stock processing.
func (this ProductType) GetSupportsStock() *bool {
	return this.GetFieldData().GetBool(ProductTypeFieldSupportsStock)
}

func (this *ProductType) SetSupportsStock(v *bool) {
	this.GetFieldData().SetBool(ProductTypeFieldSupportsStock, v)
}

func (this ProductType) GetSupportsSale() *bool {
	return this.GetFieldData().GetBool(ProductTypeFieldSupportsSale)
}

func (this *ProductType) SetSupportsSale(v *bool) {
	this.GetFieldData().SetBool(ProductTypeFieldSupportsSale, v)
}

func (this ProductType) GetSupportsPurchase() *bool {
	return this.GetFieldData().GetBool(ProductTypeFieldSupportsPurchase)
}

func (this *ProductType) SetSupportsPurchase(v *bool) {
	this.GetFieldData().SetBool(ProductTypeFieldSupportsPurchase, v)
}

func (this ProductType) GetSupportsManufacturing() *bool {
	return this.GetFieldData().GetBool(ProductTypeFieldSupportsManufacturing)
}

func (this *ProductType) SetSupportsManufacturing(v *bool) {
	this.GetFieldData().SetBool(ProductTypeFieldSupportsManufacturing, v)
}

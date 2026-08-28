package models

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	ProductVariantAttributeValueSchemaName = "inventory_product_variant_attribute_value"

	ProductVariantAttributeValueFieldId                       = basemodel.FieldId
	ProductVariantAttributeValueFieldProductVariantId         = "product_variant_id"
	ProductVariantAttributeValueFieldTemplateAttributeValueId = "template_attribute_value_id"
	ProductVariantAttributeValueFieldSalesPriceExtra           = "sales_price_extra"

	ProductVariantAttributeValueEdgeVariant                = "variant"
	ProductVariantAttributeValueEdgeTemplateAttributeValue = "template_attribute_value"
)

//go:embed product_variant_attribute_value.json
var productVariantAttributeValueSchemaJson string

func ProductVariantAttributeValueSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(productVariantAttributeValueSchemaJson)
}

type ProductVariantAttributeValue struct {
	basemodel.DynamicModelBase
}

func NewProductVariantAttributeValue() *ProductVariantAttributeValue {
	return &ProductVariantAttributeValue{basemodel.NewDynamicModel()}
}

func NewProductVariantAttributeValueFrom(src dmodel.DynamicFields) *ProductVariantAttributeValue {
	return &ProductVariantAttributeValue{basemodel.NewDynamicModel(src)}
}

func (this ProductVariantAttributeValue) GetProductVariantId() *model.Id {
	return this.GetFieldData().GetModelId(ProductVariantAttributeValueFieldProductVariantId)
}

func (this *ProductVariantAttributeValue) SetProductVariantId(v *model.Id) {
	this.GetFieldData().SetModelId(ProductVariantAttributeValueFieldProductVariantId, v)
}

func (this ProductVariantAttributeValue) GetTemplateAttributeValueId() *model.Id {
	return this.GetFieldData().GetModelId(ProductVariantAttributeValueFieldTemplateAttributeValueId)
}

func (this *ProductVariantAttributeValue) SetTemplateAttributeValueId(v *model.Id) {
	this.GetFieldData().SetModelId(ProductVariantAttributeValueFieldTemplateAttributeValueId, v)
}

func (this ProductVariantAttributeValue) GetSalesPriceExtra() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(ProductVariantAttributeValueFieldSalesPriceExtra)
}

func (this *ProductVariantAttributeValue) SetSalesPriceExtra(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(ProductVariantAttributeValueFieldSalesPriceExtra, v)
}

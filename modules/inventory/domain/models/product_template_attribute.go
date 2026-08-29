package models

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// The two template-attribute junctions let a template narrow a global attribute to a subset of its
// values. A variant references the template-scoped value, which makes "every value in a
// combination is allowed by the template" checkable rather than conventional.

const (
	ProductTemplateAttributeSchemaName = "inventory_product_template_attribute"

	ProductTemplateAttributeFieldId                = basemodel.FieldId
	ProductTemplateAttributeFieldProductTemplateId = "product_template_id"
	ProductTemplateAttributeFieldAttributeId       = "attribute_id"
	ProductTemplateAttributeFieldSequence          = "sequence"

	ProductTemplateAttributeEdgeTemplate  = "template"
	ProductTemplateAttributeEdgeAttribute = "attribute"
)

//go:embed product_template_attribute.json
var productTemplateAttributeSchemaJson string

func ProductTemplateAttributeSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(productTemplateAttributeSchemaJson)
}

type ProductTemplateAttribute struct {
	basemodel.DynamicModelBase
}

func NewProductTemplateAttribute() *ProductTemplateAttribute {
	return &ProductTemplateAttribute{basemodel.NewDynamicModel()}
}

func NewProductTemplateAttributeFrom(src dmodel.DynamicFields) *ProductTemplateAttribute {
	return &ProductTemplateAttribute{basemodel.NewDynamicModel(src)}
}

func (this ProductTemplateAttribute) GetProductTemplateId() *model.Id {
	return this.GetFieldData().GetModelId(ProductTemplateAttributeFieldProductTemplateId)
}

func (this *ProductTemplateAttribute) SetProductTemplateId(v *model.Id) {
	this.GetFieldData().SetModelId(ProductTemplateAttributeFieldProductTemplateId, v)
}

func (this ProductTemplateAttribute) GetAttributeId() *model.Id {
	return this.GetFieldData().GetModelId(ProductTemplateAttributeFieldAttributeId)
}

func (this *ProductTemplateAttribute) SetAttributeId(v *model.Id) {
	this.GetFieldData().SetModelId(ProductTemplateAttributeFieldAttributeId, v)
}

func (this ProductTemplateAttribute) GetSequence() *int32 {
	return this.GetFieldData().GetInt32(ProductTemplateAttributeFieldSequence)
}

func (this *ProductTemplateAttribute) SetSequence(v *int32) {
	this.GetFieldData().SetInt32(ProductTemplateAttributeFieldSequence, v)
}

const (
	ProductTemplateAttributeValueSchemaName = "inventory_product_template_attribute_value"

	ProductTemplateAttributeValueFieldId                  = basemodel.FieldId
	ProductTemplateAttributeValueFieldTemplateAttributeId = "template_attribute_id"
	ProductTemplateAttributeValueFieldAttributeValueId    = "attribute_value_id"
	ProductTemplateAttributeValueFieldSequence            = "sequence"
	ProductTemplateAttributeValueFieldSalesPriceExtra     = "sales_price_extra"

	ProductTemplateAttributeValueEdgeTemplateAttribute = "template_attribute"
	ProductTemplateAttributeValueEdgeAttributeValue    = "attribute_value"
)

//go:embed product_template_attribute_value.json
var productTemplateAttributeValueSchemaJson string

func ProductTemplateAttributeValueSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(productTemplateAttributeValueSchemaJson)
}

type ProductTemplateAttributeValue struct {
	basemodel.DynamicModelBase
}

func NewProductTemplateAttributeValue() *ProductTemplateAttributeValue {
	return &ProductTemplateAttributeValue{basemodel.NewDynamicModel()}
}

func NewProductTemplateAttributeValueFrom(src dmodel.DynamicFields) *ProductTemplateAttributeValue {
	return &ProductTemplateAttributeValue{basemodel.NewDynamicModel(src)}
}

func (this ProductTemplateAttributeValue) GetTemplateAttributeId() *model.Id {
	return this.GetFieldData().GetModelId(ProductTemplateAttributeValueFieldTemplateAttributeId)
}

func (this *ProductTemplateAttributeValue) SetTemplateAttributeId(v *model.Id) {
	this.GetFieldData().SetModelId(ProductTemplateAttributeValueFieldTemplateAttributeId, v)
}

func (this ProductTemplateAttributeValue) GetAttributeValueId() *model.Id {
	return this.GetFieldData().GetModelId(ProductTemplateAttributeValueFieldAttributeValueId)
}

func (this *ProductTemplateAttributeValue) SetAttributeValueId(v *model.Id) {
	this.GetFieldData().SetModelId(ProductTemplateAttributeValueFieldAttributeValueId, v)
}

func (this ProductTemplateAttributeValue) GetSalesPriceExtra() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(ProductTemplateAttributeValueFieldSalesPriceExtra)
}

func (this *ProductTemplateAttributeValue) SetSalesPriceExtra(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(ProductTemplateAttributeValueFieldSalesPriceExtra, v)
}

func (this ProductTemplateAttributeValue) GetSequence() *int32 {
	return this.GetFieldData().GetInt32(ProductTemplateAttributeValueFieldSequence)
}

func (this *ProductTemplateAttributeValue) SetSequence(v *int32) {
	this.GetFieldData().SetInt32(ProductTemplateAttributeValueFieldSequence, v)
}

package models

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	ProductAttributeValueSchemaName = "inventory_product_attribute_value"

	ProductAttributeValueFieldId          = basemodel.FieldId
	ProductAttributeValueFieldAttributeId = "attribute_id"
	ProductAttributeValueFieldCode        = "code"
	ProductAttributeValueFieldName        = "name"
	ProductAttributeValueFieldSequence    = "sequence"
	ProductAttributeValueFieldPriceExtra  = "price_extra"
	ProductAttributeValueFieldOrgId       = "org_id"

	ProductAttributeValueEdgeAttribute = "attribute"
)

//go:embed product_attribute_value.json
var productAttributeValueSchemaJson string

func ProductAttributeValueSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(productAttributeValueSchemaJson)
}

type ProductAttributeValue struct {
	basemodel.DynamicModelBase
}

func NewProductAttributeValue() *ProductAttributeValue {
	return &ProductAttributeValue{basemodel.NewDynamicModel()}
}

func NewProductAttributeValueFrom(src dmodel.DynamicFields) *ProductAttributeValue {
	return &ProductAttributeValue{basemodel.NewDynamicModel(src)}
}

func (this ProductAttributeValue) GetAttributeId() *model.Id {
	return this.GetFieldData().GetModelId(ProductAttributeValueFieldAttributeId)
}

func (this *ProductAttributeValue) SetAttributeId(v *model.Id) {
	this.GetFieldData().SetModelId(ProductAttributeValueFieldAttributeId, v)
}

func (this ProductAttributeValue) GetCode() *string {
	return this.GetFieldData().GetString(ProductAttributeValueFieldCode)
}

func (this *ProductAttributeValue) SetCode(v *string) {
	this.GetFieldData().SetString(ProductAttributeValueFieldCode, v)
}

func (this ProductAttributeValue) GetName() *model.LangJson {
	return this.GetFieldData().GetLangJson(ProductAttributeValueFieldName)
}

func (this *ProductAttributeValue) SetName(v *model.LangJson) {
	this.GetFieldData().SetLangJson(ProductAttributeValueFieldName, v)
}

func (this ProductAttributeValue) GetSequence() *int32 {
	return this.GetFieldData().GetInt32(ProductAttributeValueFieldSequence)
}

func (this *ProductAttributeValue) SetSequence(v *int32) {
	this.GetFieldData().SetInt32(ProductAttributeValueFieldSequence, v)
}

// GetPriceExtra is the surcharge for a variant carrying this value. Signed: it may also discount.
func (this ProductAttributeValue) GetPriceExtra() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(ProductAttributeValueFieldPriceExtra)
}

func (this *ProductAttributeValue) SetPriceExtra(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(ProductAttributeValueFieldPriceExtra, v)
}

func (this ProductAttributeValue) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(ProductAttributeValueFieldOrgId)
}

func (this *ProductAttributeValue) SetOrgId(v *model.Id) {
	this.GetFieldData().SetModelId(ProductAttributeValueFieldOrgId, v)
}

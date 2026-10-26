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
	ProductTemplateAttributeFieldIsEnabled         = "is_enabled"
	ProductTemplateAttributeFieldRevision          = "application_revision"
	ProductTemplateAttributeFieldModeSnapshot      = "variant_creation_snapshot"
	ProductTemplateAttributeFieldOrgId             = "org_id"

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

func (this ProductTemplateAttribute) GetIsEnabled() *bool {
	return this.GetFieldData().GetBool(ProductTemplateAttributeFieldIsEnabled)
}

func (this *ProductTemplateAttribute) SetIsEnabled(v *bool) {
	this.GetFieldData().SetBool(ProductTemplateAttributeFieldIsEnabled, v)
}

func (this ProductTemplateAttribute) GetApplicationRevision() *int32 {
	return this.GetFieldData().GetInt32(ProductTemplateAttributeFieldRevision)
}

func (this *ProductTemplateAttribute) SetApplicationRevision(v *int32) {
	this.GetFieldData().SetInt32(ProductTemplateAttributeFieldRevision, v)
}

// GetVariantCreationSnapshot is the attribute's mode when this line was applied.
func (this ProductTemplateAttribute) GetVariantCreationSnapshot() *VariantCreationMode {
	s := this.GetFieldData().GetString(ProductTemplateAttributeFieldModeSnapshot)
	if s == nil {
		return nil
	}
	return WrapVariantCreationMode(*s)
}

func (this *ProductTemplateAttribute) SetVariantCreationSnapshot(v *VariantCreationMode) {
	if v == nil {
		this.GetFieldData().SetString(ProductTemplateAttributeFieldModeSnapshot, nil)
		return
	}
	s := string(*v)
	this.GetFieldData().SetString(ProductTemplateAttributeFieldModeSnapshot, &s)
}

func (this ProductTemplateAttribute) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(ProductTemplateAttributeFieldOrgId)
}

func (this *ProductTemplateAttribute) SetOrgId(v *model.Id) {
	this.GetFieldData().SetModelId(ProductTemplateAttributeFieldOrgId, v)
}

const (
	ProductTemplateAttributeValueSchemaName = "inventory_product_template_attribute_value"

	ProductTemplateAttributeValueFieldId                  = basemodel.FieldId
	ProductTemplateAttributeValueFieldTemplateAttributeId = "template_attribute_id"
	ProductTemplateAttributeValueFieldAttributeValueId    = "attribute_value_id"
	ProductTemplateAttributeValueFieldSequence            = "sequence"
	ProductTemplateAttributeValueFieldSalesPriceExtra     = "sales_price_extra"
	ProductTemplateAttributeValueFieldIsSelected          = "is_selected"
	ProductTemplateAttributeValueFieldOrgId               = "org_id"
	ProductTemplateAttributeValueFieldAttributeValueName  = "attribute_value_name"
	ProductTemplateAttributeValueFieldAttributeValueCode  = "attribute_value_code"
	ProductTemplateAttributeValueFieldAttributeName       = "attribute_name"

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

func (this ProductTemplateAttributeValue) GetIsSelected() *bool {
	return this.GetFieldData().GetBool(ProductTemplateAttributeValueFieldIsSelected)
}

func (this *ProductTemplateAttributeValue) SetIsSelected(v *bool) {
	this.GetFieldData().SetBool(ProductTemplateAttributeValueFieldIsSelected, v)
}

func (this ProductTemplateAttributeValue) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(ProductTemplateAttributeValueFieldOrgId)
}

func (this *ProductTemplateAttributeValue) SetOrgId(v *model.Id) {
	this.GetFieldData().SetModelId(ProductTemplateAttributeValueFieldOrgId, v)
}

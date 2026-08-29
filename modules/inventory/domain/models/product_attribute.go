package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// AttributeDataType is the shape of an attribute's values. Only AttributeDataTypeOption can take
// part in a variant combination, which is built from discrete Attribute Value records.
type AttributeDataType string

const (
	AttributeDataTypeOption  = AttributeDataType("option")
	AttributeDataTypeText    = AttributeDataType("text")
	AttributeDataTypeNumber  = AttributeDataType("number")
	AttributeDataTypeDate    = AttributeDataType("date")
	AttributeDataTypeBoolean = AttributeDataType("boolean")
)

func (this AttributeDataType) String() string {
	return string(this)
}

func WrapAttributeDataType(s string) *AttributeDataType {
	t := AttributeDataType(s)
	return &t
}

// VariantCreationMode decides whether and when an attribute's values produce Product Variants.
type VariantCreationMode string

const (
	// VariantCreationModeInstant generates every valid combination as soon as values are assigned.
	VariantCreationModeInstant = VariantCreationMode("instant")

	// VariantCreationModeDynamic materializes a variant only when a combination is actually used.
	VariantCreationModeDynamic = VariantCreationMode("dynamic")

	// VariantCreationModeNever keeps the attribute out of the combination key entirely.
	VariantCreationModeNever = VariantCreationMode("never")
)

func (this VariantCreationMode) String() string {
	return string(this)
}

func WrapVariantCreationMode(s string) *VariantCreationMode {
	m := VariantCreationMode(s)
	return &m
}

// CreatesVariants reports whether this mode contributes to variant identity. Sole expression of
// the rule that NEVER is excluded from the combination key.
func (this VariantCreationMode) CreatesVariants() bool {
	return this == VariantCreationModeInstant || this == VariantCreationModeDynamic
}

const (
	ProductAttributeSchemaName = "inventory_product_attribute"

	ProductAttributeFieldId                  = basemodel.FieldId
	ProductAttributeFieldCode                = "code"
	ProductAttributeFieldName                = "name"
	ProductAttributeFieldDataType            = "data_type"
	ProductAttributeFieldVariantCreationMode = "variant_creation_mode"
	ProductAttributeFieldDisplayType         = "display_type"
	ProductAttributeFieldSequence            = "sequence"
	ProductAttributeFieldOrgId               = "org_id"
)

//go:embed product_attribute.json
var productAttributeSchemaJson string

func ProductAttributeSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(productAttributeSchemaJson)
}

type ProductAttribute struct {
	basemodel.DynamicModelBase
}

func NewProductAttribute() *ProductAttribute {
	return &ProductAttribute{basemodel.NewDynamicModel()}
}

func NewProductAttributeFrom(src dmodel.DynamicFields) *ProductAttribute {
	return &ProductAttribute{basemodel.NewDynamicModel(src)}
}

func (this ProductAttribute) GetCode() *string {
	return this.GetFieldData().GetString(ProductAttributeFieldCode)
}

func (this *ProductAttribute) SetCode(v *string) {
	this.GetFieldData().SetString(ProductAttributeFieldCode, v)
}

func (this ProductAttribute) GetName() *model.LangJson {
	return this.GetFieldData().GetLangJson(ProductAttributeFieldName)
}

func (this *ProductAttribute) SetName(v *model.LangJson) {
	this.GetFieldData().SetLangJson(ProductAttributeFieldName, v)
}

func (this ProductAttribute) GetDataType() *AttributeDataType {
	s := this.GetFieldData().GetString(ProductAttributeFieldDataType)
	if s == nil {
		return nil
	}
	return WrapAttributeDataType(*s)
}

func (this *ProductAttribute) SetDataType(v *AttributeDataType) {
	if v == nil {
		this.GetFieldData().SetString(ProductAttributeFieldDataType, nil)
		return
	}
	s := string(*v)
	this.GetFieldData().SetString(ProductAttributeFieldDataType, &s)
}

func (this ProductAttribute) GetVariantCreationMode() *VariantCreationMode {
	s := this.GetFieldData().GetString(ProductAttributeFieldVariantCreationMode)
	if s == nil {
		return nil
	}
	return WrapVariantCreationMode(*s)
}

func (this *ProductAttribute) SetVariantCreationMode(v *VariantCreationMode) {
	if v == nil {
		this.GetFieldData().SetString(ProductAttributeFieldVariantCreationMode, nil)
		return
	}
	s := string(*v)
	this.GetFieldData().SetString(ProductAttributeFieldVariantCreationMode, &s)
}

func (this ProductAttribute) GetDisplayType() *string {
	return this.GetFieldData().GetString(ProductAttributeFieldDisplayType)
}

func (this *ProductAttribute) SetDisplayType(v *string) {
	this.GetFieldData().SetString(ProductAttributeFieldDisplayType, v)
}

func (this ProductAttribute) GetSequence() *int32 {
	return this.GetFieldData().GetInt32(ProductAttributeFieldSequence)
}

func (this *ProductAttribute) SetSequence(v *int32) {
	this.GetFieldData().SetInt32(ProductAttributeFieldSequence, v)
}

func (this ProductAttribute) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(ProductAttributeFieldOrgId)
}

func (this *ProductAttribute) SetOrgId(v *model.Id) {
	this.GetFieldData().SetModelId(ProductAttributeFieldOrgId, v)
}

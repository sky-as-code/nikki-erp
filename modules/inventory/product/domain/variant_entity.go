package domain

import (
	"fmt"

	"github.com/shopspring/decimal"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

type VariantStatus string

const (
	VariantStatusDraft        = VariantStatus("draft")
	VariantStatusActive       = VariantStatus("active")
	VariantStatusDiscontinued = VariantStatus("discontinued")
)

const (
	VarAttrValRelSchemaName       = "inventory.variant_attr_val_rel"
	VarAttrValRelFieldVariantId   = "variant_id"
	VarAttrValRelFieldAttrValueId = "attribute_value_id"
)

func VariantAttrValRelSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(VarAttrValRelSchemaName).
		TableName("inventory_variant_attr_val_rel").
		ShouldBuildDb().
		ExtendBase().
		Field(
			basemodel.DefineFieldId(VarAttrValRelFieldVariantId).
				PrimaryKey(),
		).
		Field(
			basemodel.DefineFieldId(VarAttrValRelFieldAttrValueId).
				PrimaryKey(),
		)
}

const (
	VariantSchemaName = "inventory.variant"

	VarFieldId            = basemodel.FieldId
	VarFieldProductId     = "product_id"
	VarFieldName          = "name"
	VarFieldSku           = "sku"
	VarFieldBarcode       = "barcode"
	VarFieldProposedPrice = "proposed_price"
	VarFieldStatus        = "status"
	VarFieldImageUrl      = "image_url"
	VarFieldAttributes    = "attributes"

	VarEdgeProduct         = "product"
	VarEdgeAttributeValues = "attribute_values"
)

func VariantSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(VariantSchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Variant"}).
		TableName("inventory_variants").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Extend(basemodel.OrgIdModelSchemaBuilder()).
		Field(
			basemodel.DefineFieldId(VarFieldProductId).
				Label(model.LangJson{model.LanguageCodeEnUs: "Product"}).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(VarFieldName).
				Label(model.LangJson{model.LanguageCodeEnUs: "Name"}).
				DataType(dmodel.FieldDataTypeLangJson(1, model.MODEL_RULE_LONG_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(VarFieldSku).
				Label(model.LangJson{model.LanguageCodeEnUs: "SKU"}).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_TINY_NAME_LENGTH)).
				Unique(),
		).
		Field(
			dmodel.DefineField().
				Name(VarFieldBarcode).
				Label(model.LangJson{model.LanguageCodeEnUs: "Barcode"}).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_TINY_NAME_LENGTH)),
		).
		Field(
			dmodel.DefineField().
				Name(VarFieldProposedPrice).
				Label(model.LangJson{model.LanguageCodeEnUs: "Proposed Price"}).
				DataType(dmodel.FieldDataTypeDecimal("0", fmt.Sprint(model.MODEL_RULE_CURRENCY_MAX), model.MODEL_RULE_CURRENCY_SCALE)),
		).
		Field(
			dmodel.DefineField().
				Name(VarFieldStatus).
				Label(model.LangJson{model.LanguageCodeEnUs: "Status"}).
				DataType(dmodel.FieldDataTypeEnumString([]string{
					string(VariantStatusDraft),
					string(VariantStatusActive),
					string(VariantStatusDiscontinued),
				})).
				Default(string(VariantStatusActive)),
		).
		Field(
			dmodel.DefineField().
				Name(VarFieldImageUrl).
				Label(model.LangJson{model.LanguageCodeEnUs: "Image URL"}).
				DataType(dmodel.FieldDataTypeUrl()),
		).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		EdgeTo(
			dmodel.Edge(VarEdgeProduct).
				Label(model.LangJson{model.LanguageCodeEnUs: "Product"}).
				ManyToOne(ProductSchemaName, dmodel.DynamicFields{
					VarFieldProductId: basemodel.FieldId,
				}).
				OnDelete(dmodel.RelationCascadeCascade),
		).
		EdgeTo(
			dmodel.Edge(VarEdgeAttributeValues).
				Label(model.LangJson{model.LanguageCodeEnUs: "Attribute Values"}).
				ManyToMany(AttributeValueSchemaName, VarAttrValRelSchemaName, "variant").
				OnDelete(dmodel.RelationCascadeCascade),
		)
}

type Variant struct {
	basemodel.DynamicModelBase
}

func NewVariant() *Variant {
	return &Variant{basemodel.NewDynamicModel()}
}

func NewVariantFrom(src dmodel.DynamicFields) *Variant {
	return &Variant{basemodel.NewDynamicModel(src)}
}

func (this Variant) GetProductId() *model.Id {
	return this.GetFieldData().GetModelId(VarFieldProductId)
}

func (this *Variant) SetProductId(v *model.Id) {
	this.GetFieldData().SetModelId(VarFieldProductId, v)
}

func (this Variant) GetName() *model.LangJson {
	v := this.GetFieldData().GetAny(VarFieldName)
	if v == nil {
		return nil
	}
	if langJson, ok := v.(model.LangJson); ok {
		return &langJson
	}
	return nil
}

func (this *Variant) SetName(v *model.LangJson) {
	if v == nil {
		this.GetFieldData().SetAny(VarFieldName, nil)
		return
	}
	this.GetFieldData().SetAny(VarFieldName, *v)
}

func (this Variant) GetSku() *string {
	return this.GetFieldData().GetString(VarFieldSku)
}

func (this *Variant) SetSku(v *string) {
	this.GetFieldData().SetString(VarFieldSku, v)
}

func (this Variant) GetBarcode() *string {
	return this.GetFieldData().GetString(VarFieldBarcode)
}

func (this *Variant) SetBarcode(v *string) {
	this.GetFieldData().SetString(VarFieldBarcode, v)
}

func (this Variant) GetProposedPrice() *decimal.Decimal {
	v := this.GetFieldData().GetAny(VarFieldProposedPrice)
	if v == nil {
		return nil
	}
	d := v.(decimal.Decimal)
	return &d
}

func (this *Variant) SetProposedPrice(v *decimal.Decimal) {
	if v == nil {
		this.GetFieldData().SetAny(VarFieldProposedPrice, nil)
		return
	}
	this.GetFieldData().SetAny(VarFieldProposedPrice, *v)
}

func (this Variant) GetStatus() *VariantStatus {
	s := this.GetFieldData().GetString(VarFieldStatus)
	if s == nil {
		return nil
	}
	st := VariantStatus(*s)
	return &st
}

func (this *Variant) SetStatus(v *VariantStatus) {
	if v == nil {
		this.GetFieldData().SetString(VarFieldStatus, nil)
		return
	}
	s := string(*v)
	this.GetFieldData().SetString(VarFieldStatus, &s)
}

func (this Variant) GetImageUrl() *string {
	return this.GetFieldData().GetString(VarFieldImageUrl)
}

func (this *Variant) SetImageUrl(v *string) {
	this.GetFieldData().SetString(VarFieldImageUrl, v)
}

func (this Variant) GetAttributes() map[string]any {
	v := this.GetFieldData().GetAny(VarFieldAttributes)
	if v == nil {
		return nil
	}
	if attrMap, ok := v.(map[string]any); ok {
		return attrMap
	}
	return nil
}

func (this *Variant) SetAttributes(v map[string]any) {
	this.GetFieldData().SetAny(VarFieldAttributes, v)
}

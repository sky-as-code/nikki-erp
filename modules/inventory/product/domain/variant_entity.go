package domain

import (
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

func (this VariantStatus) String() string {
	return string(this)
}

func WrapVariantStatus(s string) *VariantStatus {
	st := VariantStatus(s)
	return &st
}

const (
	VarAttrValRelSchemaName       = "inventory.variant_attr_val_rel"
	VarAttrValRelFieldVariantId   = "variant_id"
	VarAttrValRelFieldAttrValueId = "attribute_value_id"
)

func VariantAttrValRelSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(VarAttrValRelSchemaName).
		TableName("invent_variant_attr_val_rel").
		ShouldBuildDb().
		Field(
			dmodel.DefineField().
				Name(VarAttrValRelFieldVariantId).
				DataType(dmodel.FieldDataTypeUlid()).
				PrimaryKey(),
		).
		Field(
			dmodel.DefineField().
				Name(VarAttrValRelFieldAttrValueId).
				DataType(dmodel.FieldDataTypeUlid()).
				PrimaryKey(),
		)
}

const (
	VariantSchemaName = "inventory.variant"

	VarFieldId            = basemodel.FieldId
	VarFieldProductId     = "product_id"
	VarFieldOrgId         = "org_id"
	VarFieldName          = "name"
	VarFieldSku           = "sku"
	VarFieldBarcode       = "barcode"
	VarFieldProposedPrice = "proposed_price"
	VarFieldStatus        = "status"
	VarFieldImageUrl      = "image_url"

	VarEdgeProduct         = "product"
	VarEdgeAttributeValues = "attribute_values"
)

func VariantSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(VariantSchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Variant"}).
		TableName("invent_variants").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			dmodel.DefineField().
				Name(VarFieldProductId).
				Label(model.LangJson{model.LanguageCodeEnUs: "Product"}).
				DataType(dmodel.FieldDataTypeUlid()).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(VarFieldOrgId).
				Label(model.LangJson{model.LanguageCodeEnUs: "Organization"}).
				DataType(dmodel.FieldDataTypeUlid()).
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
				DataType(dmodel.FieldDataTypeDecimal("0", "9999999999.9999", 4)),
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
	fields dmodel.DynamicFields
}

func NewVariant() *Variant {
	return &Variant{fields: make(dmodel.DynamicFields)}
}

func NewVariantFrom(src dmodel.DynamicFields) *Variant {
	return &Variant{fields: src}
}

func (this Variant) GetFieldData() dmodel.DynamicFields {
	return this.fields
}

func (this *Variant) SetFieldData(data dmodel.DynamicFields) {
	this.fields = data
}

func (this Variant) GetId() *model.Id {
	return this.fields.GetModelId(basemodel.FieldId)
}

func (this *Variant) SetId(v *model.Id) {
	this.fields.SetModelId(basemodel.FieldId, v)
}

func (this Variant) GetProductId() *model.Id {
	return this.fields.GetModelId(VarFieldProductId)
}

func (this *Variant) SetProductId(v *model.Id) {
	this.fields.SetModelId(VarFieldProductId, v)
}

func (this Variant) GetName() *model.LangJson {
	v := this.fields.GetAny(VarFieldName)
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
		this.fields.SetAny(VarFieldName, nil)
		return
	}
	this.fields.SetAny(VarFieldName, *v)
}

func (this Variant) GetSku() *string {
	return this.fields.GetString(VarFieldSku)
}

func (this *Variant) SetSku(v *string) {
	this.fields.SetString(VarFieldSku, v)
}

func (this Variant) GetBarcode() *string {
	return this.fields.GetString(VarFieldBarcode)
}

func (this *Variant) SetBarcode(v *string) {
	this.fields.SetString(VarFieldBarcode, v)
}

func (this Variant) GetProposedPrice() *decimal.Decimal {
	v := this.fields.GetAny(VarFieldProposedPrice)
	if v == nil {
		return nil
	}
	d := v.(decimal.Decimal)
	return &d
}

func (this *Variant) SetProposedPrice(v *decimal.Decimal) {
	if v == nil {
		this.fields.SetAny(VarFieldProposedPrice, nil)
		return
	}
	this.fields.SetAny(VarFieldProposedPrice, *v)
}

func (this Variant) GetStatus() *VariantStatus {
	s := this.fields.GetString(VarFieldStatus)
	if s == nil {
		return nil
	}
	st := VariantStatus(*s)
	return &st
}

func (this *Variant) SetStatus(v *VariantStatus) {
	if v == nil {
		this.fields.SetString(VarFieldStatus, nil)
		return
	}
	s := string(*v)
	this.fields.SetString(VarFieldStatus, &s)
}

func (this Variant) GetImageUrl() *string {
	return this.fields.GetString(VarFieldImageUrl)
}

func (this *Variant) SetImageUrl(v *string) {
	this.fields.SetString(VarFieldImageUrl, v)
}

func (this Variant) GetOrgId() *model.Id {
	return this.fields.GetModelId(VarFieldOrgId)
}

func (this *Variant) SetOrgId(v *model.Id) {
	this.fields.SetModelId(VarFieldOrgId, v)
}

func (this Variant) GetEtag() *model.Etag {
	return this.fields.GetEtag(basemodel.FieldEtag)
}

func (this *Variant) SetEtag(v *model.Etag) {
	this.fields.SetEtag(basemodel.FieldEtag, v)
}

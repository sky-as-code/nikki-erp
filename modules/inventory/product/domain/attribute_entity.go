package domain

import (
	"math"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/json"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

type AttributeDataType string

const (
	AttributeDataTypeText      = AttributeDataType("text")
	AttributeDataTypeNumber    = AttributeDataType("number")
	AttributeDataTypeBoolean   = AttributeDataType("boolean")
	AttributeDataTypeReference = AttributeDataType("reference")
)

func (this AttributeDataType) String() string {
	return string(this)
}

const (
	AttributeSchemaName = "inventory.attribute"

	AttrFieldId               = basemodel.FieldId
	AttrFieldCodeName         = "code_name"
	AttrFieldDisplayName      = "display_name"
	AttrFieldSortIndex        = "sort_index"
	AttrFieldDataType         = "data_type"
	AttrFieldIsRequired       = "is_required"
	AttrFieldIsEnum           = "is_enum"
	AttrFieldEnumValueSort    = "enum_value_sort"
	AttrFieldEnumValue        = "enum_value"
	AttrFieldAttributeGroupId = "attribute_group_id"
	AttrFieldProductId        = "product_id"

	AttrEdgeProduct = "product"
	AttrEdgeGroup   = "group"
	AttrEdgeValues  = "values"
)

func AttributeSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(AttributeSchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Attribute"}).
		TableName("invent_attributes").
		CompositeUnique(AttrFieldCodeName, AttrFieldProductId).
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			dmodel.DefineField().
				Name(AttrFieldCodeName).
				Label(model.LangJson{model.LanguageCodeEnUs: "Code Name"}).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_TINY_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(AttrFieldDisplayName).
				Label(model.LangJson{model.LanguageCodeEnUs: "Display Name"}).
				DataType(dmodel.FieldDataTypeLangJson(1, model.MODEL_RULE_LONG_NAME_LENGTH)),
		).
		Field(
			dmodel.DefineField().
				Name(AttrFieldSortIndex).
				Label(model.LangJson{model.LanguageCodeEnUs: "Sort Index"}).
				DataType(dmodel.FieldDataTypeInt64(0, math.MaxInt16)).
				Default(0),
		).
		Field(
			dmodel.DefineField().
				Name(AttrFieldDataType).
				Label(model.LangJson{model.LanguageCodeEnUs: "Data Type"}).
				DataType(dmodel.FieldDataTypeEnumString([]string{
					string(AttributeDataTypeText),
					string(AttributeDataTypeNumber),
					string(AttributeDataTypeBoolean),
					string(AttributeDataTypeReference),
				})).
				RequiredForCreate().
				Default(string(AttributeDataTypeText)),
		).
		Field(
			dmodel.DefineField().
				Name(AttrFieldIsRequired).
				Label(model.LangJson{model.LanguageCodeEnUs: "Is Required"}).
				DataType(dmodel.FieldDataTypeBoolean()).
				Default(false),
		).
		Field(
			dmodel.DefineField().
				Name(AttrFieldIsEnum).
				Label(model.LangJson{model.LanguageCodeEnUs: "Is Enum"}).
				DataType(dmodel.FieldDataTypeBoolean()).
				Default(false),
		).
		Field(
			dmodel.DefineField().
				Name(AttrFieldEnumValueSort).
				Label(model.LangJson{model.LanguageCodeEnUs: "Enum Value Sort"}).
				DataType(dmodel.FieldDataTypeBoolean()).
				Default(false),
		).
		Field(
			dmodel.DefineField().
				Name(AttrFieldEnumValue).
				Label(model.LangJson{model.LanguageCodeEnUs: "Enum Values"}).
				DataType(dmodel.FieldDataTypeLangJson(1, model.MODEL_RULE_LONG_NAME_LENGTH).ArrayType()),
		).
		Field(
			dmodel.DefineField().
				Name(AttrFieldAttributeGroupId).
				Label(model.LangJson{model.LanguageCodeEnUs: "Attribute Group"}).
				DataType(dmodel.FieldDataTypeUlid()),
		).
		Field(
			dmodel.DefineField().
				Name(AttrFieldProductId).
				Label(model.LangJson{model.LanguageCodeEnUs: "Product"}).
				DataType(dmodel.FieldDataTypeUlid()).
				RequiredForCreate(),
		).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		EdgeTo(
			dmodel.Edge(AttrEdgeProduct).
				Label(model.LangJson{model.LanguageCodeEnUs: "Product"}).
				ManyToOne(ProductSchemaName, dmodel.DynamicFields{
					AttrFieldProductId: basemodel.FieldId,
				}).
				OnDelete(dmodel.RelationCascadeCascade),
		).
		EdgeTo(
			dmodel.Edge(AttrEdgeGroup).
				Label(model.LangJson{model.LanguageCodeEnUs: "Attribute Group"}).
				ManyToOne(AttributeGroupSchemaName, dmodel.DynamicFields{
					AttrFieldAttributeGroupId: basemodel.FieldId,
				}).
				OnDelete(dmodel.RelationCascadeSetNull),
		).
		EdgeFrom(
			dmodel.Edge(AttrEdgeValues).
				Label(model.LangJson{model.LanguageCodeEnUs: "Attribute Values"}).
				Existing(AttributeValueSchemaName, AttrValEdgeAttribute),
		)
}

type Attribute struct {
	fields dmodel.DynamicFields
}

func NewAttribute() *Attribute {
	return &Attribute{fields: make(dmodel.DynamicFields)}
}

func NewAttributeFrom(src dmodel.DynamicFields) *Attribute {
	return &Attribute{fields: src}
}

func (this Attribute) GetFieldData() dmodel.DynamicFields {
	return this.fields
}

func (this *Attribute) SetFieldData(data dmodel.DynamicFields) {
	this.fields = data
}

func (this Attribute) GetId() *model.Id {
	return this.fields.GetModelId(basemodel.FieldId)
}

func (this *Attribute) SetId(v *model.Id) {
	this.fields.SetModelId(basemodel.FieldId, v)
}

func (this Attribute) GetCodeName() *string {
	return this.fields.GetString(AttrFieldCodeName)
}

func (this *Attribute) SetCodeName(v *string) {
	this.fields.SetString(AttrFieldCodeName, v)
}

func (this Attribute) GetDisplayName() *model.LangJson {
	v := this.fields.GetAny(AttrFieldDisplayName)
	if v == nil {
		return nil
	}
	// From DB read: value is a JSON string; unmarshal it.
	if s, ok := v.(string); ok && s != "" {
		var lj model.LangJson
		if err := json.Unmarshal([]byte(s), &lj); err == nil {
			return &lj
		}
	}
	return nil
}

func (this *Attribute) SetDisplayName(v *model.LangJson) {
	if v == nil {
		this.fields.SetAny(AttrFieldDisplayName, nil)
		return
	}
	this.fields.SetAny(AttrFieldDisplayName, *v)
}

func (this Attribute) GetSortIndex() *int64 {
	return this.fields.GetInt64(AttrFieldSortIndex)
}

func (this *Attribute) SetSortIndex(v *int64) {
	this.fields.SetInt64(AttrFieldSortIndex, v)
}

func (this Attribute) GetDataType() *AttributeDataType {
	s := this.fields.GetString(AttrFieldDataType)
	if s == nil {
		return nil
	}
	dt := AttributeDataType(*s)
	return &dt
}

func (this *Attribute) SetDataType(v *AttributeDataType) {
	if v == nil {
		this.fields.SetString(AttrFieldDataType, nil)
		return
	}
	s := string(*v)
	this.fields.SetString(AttrFieldDataType, &s)
}

func (this Attribute) GetIsRequired() *bool {
	return this.fields.GetBool(AttrFieldIsRequired)
}

func (this *Attribute) SetIsRequired(v *bool) {
	this.fields.SetBool(AttrFieldIsRequired, v)
}

func (this Attribute) GetIsEnum() *bool {
	return this.fields.GetBool(AttrFieldIsEnum)
}

func (this *Attribute) SetIsEnum(v *bool) {
	this.fields.SetBool(AttrFieldIsEnum, v)
}

func (this Attribute) GetEnumValueSort() *bool {
	return this.fields.GetBool(AttrFieldEnumValueSort)
}

func (this *Attribute) SetEnumValueSort(v *bool) {
	this.fields.SetBool(AttrFieldEnumValueSort, v)
}

func (this Attribute) GetEnumValue() []model.LangJson {
	v := this.fields.GetAny(AttrFieldEnumValue)
	if v == nil {
		return nil
	}
	// From DB read: value is a JSON string; unmarshal it.
	if s, ok := v.(string); ok {
		var result []model.LangJson
		if err := json.Unmarshal([]byte(s), &result); err != nil {
			return nil
		}
		return result
	}
	// From in-memory validation: value is []any{model.LangJson{...}, ...}
	if items, ok := v.([]any); ok {
		result := make([]model.LangJson, 0, len(items))
		for _, item := range items {
			if lj, ok := item.(model.LangJson); ok {
				result = append(result, lj)
			} else if ljMap, ok := item.(map[string]string); ok {
				result = append(result, model.LangJson(ljMap))
			} else if ljMap, ok := item.(map[string]interface{}); ok {
				// Convert map[string]interface{} to LangJson (from JSON unmarshaling)
				converted := make(model.LangJson)
				for k, val := range ljMap {
					if str, ok := val.(string); ok {
						converted[k] = str
					}
				}
				result = append(result, converted)
			} else if jsonStr, ok := item.(string); ok {
				// Handle case where each item is a JSON string
				var lj model.LangJson
				if err := json.Unmarshal([]byte(jsonStr), &lj); err == nil {
					result = append(result, lj)
				}
			}
		}
		return result
	}
	return nil
}

func (this *Attribute) SetEnumValue(v []model.LangJson) {
	if v == nil {
		this.fields.SetAny(AttrFieldEnumValue, nil)
		return
	}
	anySlice := make([]any, len(v))
	for i, lj := range v {
		anySlice[i] = lj
	}
	this.fields.SetAny(AttrFieldEnumValue, anySlice)
}

func (this Attribute) GetAttributeGroupId() *model.Id {
	return this.fields.GetModelId(AttrFieldAttributeGroupId)
}

func (this *Attribute) SetAttributeGroupId(v *model.Id) {
	this.fields.SetModelId(AttrFieldAttributeGroupId, v)
}

func (this Attribute) GetProductId() *model.Id {
	return this.fields.GetModelId(AttrFieldProductId)
}

func (this *Attribute) SetProductId(v *model.Id) {
	this.fields.SetModelId(AttrFieldProductId, v)
}

func (this Attribute) GetEtag() *model.Etag {
	return this.fields.GetEtag(basemodel.FieldEtag)
}

func (this *Attribute) SetEtag(v *model.Etag) {
	this.fields.SetEtag(basemodel.FieldEtag, v)
}

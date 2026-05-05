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
	AttributeDataTypeBoolean = AttributeDataType("boolean")
	AttributeDataTypeNumber  = AttributeDataType("number")
	AttributeDataTypeUrl     = AttributeDataType("url")
	AttributeDataTypeText    = AttributeDataType("text")
	AttributeDataTypeUnit    = AttributeDataType("unit")
)

const (
	AttributeResourceCode = "inventory_attribute"
	AttributeAuthScope    = "org"

	AttributeActionCreate = "create"
	AttributeActionDelete = "delete"
	AttributeActionUpdate = "update"
	AttributeActionView   = "view"
)

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
	AttrFieldEnumValueText    = "enum_value_text"
	AttrFieldEnumValueNumber  = "enum_value_number"
	AttrFieldAttributeGroupId = "attribute_group_id"
	AttrFieldProductId        = "product_id"

	AttrEdgeProduct = "product"
	AttrEdgeGroup   = "group"
	AttrEdgeValues  = "values"
)

func AttributeSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(AttributeSchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Attribute"}).
		TableName("inventory_attributes").
		CompositeUnique(AttrFieldCodeName, AttrFieldProductId).
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			dmodel.DefineField().
				Name(AttrFieldCodeName).
				Label(model.LangJson{model.LanguageCodeEnUs: "Code Name"}).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_TINY_NAME_LENGTH)).
				Unique().
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
					string(AttributeDataTypeBoolean),
					string(AttributeDataTypeNumber),
					string(AttributeDataTypeUrl),
					string(AttributeDataTypeText),
					string(AttributeDataTypeUnit),
				})).
				RequiredForCreate(),
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
				Name(AttrFieldEnumValueText).
				Label(model.LangJson{model.LanguageCodeEnUs: "Enum Values Text"}).
				DataType(dmodel.FieldDataTypeLangJson(1, model.MODEL_RULE_LONG_NAME_LENGTH).ArrayType()),
		).
		Field(
			dmodel.DefineField().
				Name(AttrFieldEnumValueNumber).
				Label(model.LangJson{model.LanguageCodeEnUs: "Enum Values Number"}).
				DataType(dmodel.FieldDataTypeInt64(0, math.MaxInt64).ArrayType()),
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
		Extend(basemodel.ArchivableModelSchemaBuilder()).
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
	basemodel.DynamicModelBase
}

func NewAttribute() *Attribute {
	return &Attribute{basemodel.NewDynamicModel()}
}

func NewAttributeFrom(src dmodel.DynamicFields) *Attribute {
	return &Attribute{basemodel.NewDynamicModel(src)}
}

func (this Attribute) GetCodeName() *string {
	return this.GetFieldData().GetString(AttrFieldCodeName)
}

func (this *Attribute) SetCodeName(v *string) {
	this.GetFieldData().SetString(AttrFieldCodeName, v)
}

func (this Attribute) GetDisplayName() *model.LangJson {
	v := this.GetFieldData().GetAny(AttrFieldDisplayName)
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
		this.GetFieldData().SetAny(AttrFieldDisplayName, nil)
		return
	}
	this.GetFieldData().SetAny(AttrFieldDisplayName, *v)
}

func (this Attribute) GetSortIndex() *int64 {
	return this.GetFieldData().GetInt64(AttrFieldSortIndex)
}

func (this *Attribute) SetSortIndex(v *int64) {
	this.GetFieldData().SetInt64(AttrFieldSortIndex, v)
}

func (this Attribute) GetDataType() *AttributeDataType {
	s := this.GetFieldData().GetString(AttrFieldDataType)
	if s == nil {
		return nil
	}
	dt := AttributeDataType(*s)
	return &dt
}

func (this *Attribute) SetDataType(v *AttributeDataType) {
	if v == nil {
		this.GetFieldData().SetString(AttrFieldDataType, nil)
		return
	}
	s := string(*v)
	this.GetFieldData().SetString(AttrFieldDataType, &s)
}

func (this Attribute) GetIsRequired() *bool {
	return this.GetFieldData().GetBool(AttrFieldIsRequired)
}

func (this *Attribute) SetIsRequired(v *bool) {
	this.GetFieldData().SetBool(AttrFieldIsRequired, v)
}

func (this Attribute) GetIsEnum() *bool {
	return this.GetFieldData().GetBool(AttrFieldIsEnum)
}

func (this *Attribute) SetIsEnum(v *bool) {
	this.GetFieldData().SetBool(AttrFieldIsEnum, v)
}

func (this Attribute) GetEnumValueSort() *bool {
	return this.GetFieldData().GetBool(AttrFieldEnumValueSort)
}

func (this *Attribute) SetEnumValueSort(v *bool) {
	this.GetFieldData().SetBool(AttrFieldEnumValueSort, v)
}

func (this Attribute) GetEnumValueText() []model.LangJson {
	v := this.GetFieldData().GetAny(AttrFieldEnumValueText)
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

func (this *Attribute) SetEnumValueText(v []model.LangJson) {
	if v == nil {
		this.GetFieldData().SetAny(AttrFieldEnumValueText, nil)
		return
	}
	anySlice := make([]any, len(v))
	for i, lj := range v {
		anySlice[i] = lj
	}
	this.GetFieldData().SetAny(AttrFieldEnumValueText, anySlice)
}

func (this Attribute) GetEnumValueNumber() []int64 {
	v := this.GetFieldData().GetAny(AttrFieldEnumValueNumber)
	if v == nil {
		return nil
	}
	// From DB read: value is a JSON string; unmarshal it.
	if s, ok := v.(string); ok {
		var result []int64
		if err := json.Unmarshal([]byte(s), &result); err != nil {
			return nil
		}
		return result
	}
	// From in-memory validation: value is []any{int64, ...} or []interface{}
	if items, ok := v.([]any); ok {
		result := make([]int64, 0, len(items))
		for _, item := range items {
			// Try direct int64
			if num, ok := item.(int64); ok {
				result = append(result, num)
			} else if num, ok := item.(int); ok {
				result = append(result, int64(num))
			} else if num, ok := item.(float64); ok {
				// JSON numbers are unmarshaled as float64
				result = append(result, int64(num))
			} else if num, ok := item.(int32); ok {
				result = append(result, int64(num))
			}
		}
		return result
	}
	return nil
}

func (this *Attribute) SetEnumValueNumber(v []int64) {
	if v == nil {
		this.GetFieldData().SetAny(AttrFieldEnumValueNumber, nil)
		return
	}
	anySlice := make([]any, len(v))
	for i, num := range v {
		anySlice[i] = num
	}
	this.GetFieldData().SetAny(AttrFieldEnumValueNumber, anySlice)
}

func (this Attribute) GetAttributeGroupId() *model.Id {
	return this.GetFieldData().GetModelId(AttrFieldAttributeGroupId)
}

func (this *Attribute) SetAttributeGroupId(v *model.Id) {
	this.GetFieldData().SetModelId(AttrFieldAttributeGroupId, v)
}

func (this Attribute) GetProductId() *model.Id {
	return this.GetFieldData().GetModelId(AttrFieldProductId)
}

func (this *Attribute) SetProductId(v *model.Id) {
	this.GetFieldData().SetModelId(AttrFieldProductId, v)
}

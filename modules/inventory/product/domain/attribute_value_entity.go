package domain

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	AttributeValueSchemaName = "inventory.attribute_value"

	AttrValFieldId           = basemodel.FieldId
	AttrValFieldAttributeId  = "attribute_id"
	AttrValFieldValueText    = "value_text"
	AttrValFieldValueInteger = "value_integer"
	AttrValFieldValueDecimal = "value_decimal"
	AttrValFieldValueBool    = "value_bool"
	AttrValFieldValueRef     = "value_ref"
	AttrValFieldVariantCount = "variant_count"

	AttrValEdgeAttribute = "attribute"
	AttrValEdgeVariants  = "variants"
)

func AttributeValueSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(AttributeValueSchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Attribute Value"}).
		TableName("inventory_attribute_values").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			basemodel.DefineFieldId(AttrValFieldAttributeId).
				RequiredForCreate(),
		).
		ExclusiveFields(AttrValFieldValueText, AttrValFieldValueDecimal, AttrValFieldValueInteger, AttrValFieldValueBool, AttrValFieldValueRef).
		Field(
			dmodel.DefineField().
				Name(AttrValFieldValueText).
				Label(model.LangJson{model.LanguageCodeEnUs: "Text Value"}).
				DataType(dmodel.FieldDataTypeLangJson(0, model.MODEL_RULE_LONG_NAME_LENGTH)),
		).
		Field(
			dmodel.DefineField().
				Name(AttrValFieldValueDecimal).
				Label(model.LangJson{model.LanguageCodeEnUs: "Decimal Value"}).
				DataType(dmodel.FieldDataTypeDecimal("0", fmt.Sprint(model.MODEL_RULE_CURRENCY_MAX), model.MODEL_RULE_CURRENCY_SCALE)),
		).
		Field(
			dmodel.DefineField().
				Name(AttrValFieldValueInteger).
				Label(model.LangJson{model.LanguageCodeEnUs: "Integer Value"}).
				DataType(dmodel.FieldDataTypeInt64(0, math.MaxInt64)),
		).
		Field(
			dmodel.DefineField().
				Name(AttrValFieldValueBool).
				Label(model.LangJson{model.LanguageCodeEnUs: "Boolean Value"}).
				DataType(dmodel.FieldDataTypeBoolean()),
		).
		Field(
			dmodel.DefineField().
				Name(AttrValFieldValueRef).
				Label(model.LangJson{model.LanguageCodeEnUs: "Reference Value"}).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_LONG_NAME_LENGTH)),
		).
		Field(
			dmodel.DefineField().
				Name(AttrValFieldVariantCount).
				Label(model.LangJson{model.LanguageCodeEnUs: "Variant Count"}).
				DataType(dmodel.FieldDataTypeInt64(0, math.MaxInt16)).
				Default(0),
		).
		EdgeTo(
			dmodel.Edge(AttrValEdgeAttribute).
				Label(model.LangJson{model.LanguageCodeEnUs: "Attribute"}).
				ManyToOne(AttributeSchemaName, dmodel.DynamicFields{
					AttrValFieldAttributeId: basemodel.FieldId,
				}).
				OnDelete(dmodel.RelationCascadeCascade),
		).
		EdgeTo(
			dmodel.Edge(AttrValEdgeVariants).
				Label(model.LangJson{model.LanguageCodeEnUs: "Variants"}).
				ManyToMany(VariantSchemaName, VarAttrValRelSchemaName, "attribute_value").
				OnDelete(dmodel.RelationCascadeCascade),
		)
}

type AttributeValue struct {
	basemodel.DynamicModelBase
}

func NewAttributeValue() *AttributeValue {
	return &AttributeValue{basemodel.NewDynamicModel()}
}

func NewAttributeValueFrom(src dmodel.DynamicFields) *AttributeValue {
	return &AttributeValue{basemodel.NewDynamicModel(src)}
}

func (this AttributeValue) GetAttributeId() *model.Id {
	return this.GetFieldData().GetModelId(AttrValFieldAttributeId)
}

func (this *AttributeValue) SetAttributeId(v *model.Id) {
	this.GetFieldData().SetModelId(AttrValFieldAttributeId, v)
}

func (this AttributeValue) GetValueText() *model.LangJson {
	v := this.GetFieldData().GetAny(AttrValFieldValueText)
	if v == nil {
		return nil
	}
	lj := v.(model.LangJson)
	return &lj
}

func (this *AttributeValue) SetValueText(v *model.LangJson) {
	if v == nil {
		this.GetFieldData().SetAny(AttrValFieldValueText, nil)
		return
	}
	this.GetFieldData().SetAny(AttrValFieldValueText, *v)
}

func (this AttributeValue) GetValueDecimal() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(AttrValFieldValueDecimal)
}

func (this *AttributeValue) SetValueDecimal(v *string) {
	this.GetFieldData().SetDecimalStr(AttrValFieldValueDecimal, v)
}

func (this AttributeValue) GetValueInteger() *int64 {
	return this.GetFieldData().GetInt64(AttrValFieldValueInteger)
}

func (this *AttributeValue) SetValueInteger(v *int64) {
	this.GetFieldData().SetInt64(AttrValFieldValueInteger, v)
}

func (this AttributeValue) GetValueBool() *bool {
	return this.GetFieldData().GetBool(AttrValFieldValueBool)
}

func (this *AttributeValue) SetValueBool(v *bool) {
	this.GetFieldData().SetBool(AttrValFieldValueBool, v)
}

func (this AttributeValue) GetValueRef() *string {
	return this.GetFieldData().GetString(AttrValFieldValueRef)
}

func (this *AttributeValue) SetValueRef(v *string) {
	this.GetFieldData().SetString(AttrValFieldValueRef, v)
}

func (this AttributeValue) GetVariantCount() *int64 {
	return this.GetFieldData().GetInt64(AttrValFieldVariantCount)
}

func (this *AttributeValue) SetVariantCount(v *int64) {
	this.GetFieldData().SetInt64(AttrValFieldVariantCount, v)
}

// ExpectedValueFieldForDataType returns the field name that should be set for the given data type.
func ExpectedValueFieldForDataType(dataType AttributeDataType) string {
	switch dataType {
	case AttributeDataTypeText:
		return AttrValFieldValueText
	case AttributeDataTypeNumber:
		return AttrValFieldValueDecimal
	case AttributeDataTypeBoolean:
		return AttrValFieldValueBool
	case AttributeDataTypeUnit:
		return AttrValFieldValueRef
	case AttributeDataTypeUrl:
		return AttrValFieldValueInteger
	default:
		return ""
	}
}

// GetValue returns the field name and concrete value of whichever value field is set.
// Returns ("", nil) if no value field is set.
func (this AttributeValue) GetValue() (string, any) {
	if v := this.GetValueText(); v != nil {
		return AttrValFieldValueText, *v
	}
	if v := this.GetValueDecimal(); v != nil {
		return AttrValFieldValueDecimal, *v
	}
	if v := this.GetValueInteger(); v != nil {
		return AttrValFieldValueInteger, *v
	}
	if v := this.GetValueBool(); v != nil {
		return AttrValFieldValueBool, *v
	}
	if v := this.GetValueRef(); v != nil {
		return AttrValFieldValueRef, *v
	}
	return "", nil
}

// SetValueFromRaw converts the raw input and sets the appropriate value field for the given data type.
// Returns an error if the raw value cannot be converted to the expected Go type.
func (this *AttributeValue) SetValueFromRaw(dataType AttributeDataType, value any) error {
	switch dataType {
	case AttributeDataTypeNumber:
		switch v := value.(type) {
		case string:
			this.SetValueDecimal(&v)
		case float64:
			strVal := fmt.Sprintf("%g", v)
			this.SetValueDecimal(&strVal)
		case int:
			strVal := fmt.Sprintf("%d", v)
			this.SetValueDecimal(&strVal)
		case int64:
			strVal := fmt.Sprintf("%d", v)
			this.SetValueDecimal(&strVal)
		default:
			return fmt.Errorf("value must be a number")
		}

	case AttributeDataTypeUrl:
		switch v := value.(type) {
		case int64:
			this.SetValueInteger(&v)
		case float64:
			valueInt := int64(v)
			this.SetValueInteger(&valueInt)
		case int:
			valueInt := int64(v)
			this.SetValueInteger(&valueInt)
		default:
			return fmt.Errorf("value must be an integer")
		}

	case AttributeDataTypeBoolean:
		v, ok := value.(bool)
		if !ok {
			return fmt.Errorf("value must be a boolean")
		}
		this.SetValueBool(&v)

	case AttributeDataTypeText:
		bytes, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("value must be valid JSON")
		}
		var langJson model.LangJson
		if err := json.Unmarshal(bytes, &langJson); err != nil {
			return fmt.Errorf("value must be a valid language JSON structure")
		}
		this.SetValueText(&langJson)

	case AttributeDataTypeUnit:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("value must be a string reference")
		}
		this.SetValueRef(&v)

	default:
		return fmt.Errorf("unsupported attribute data type: %s", dataType)
	}
	return nil
}

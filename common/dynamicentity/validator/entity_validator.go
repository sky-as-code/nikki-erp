package validator

import (
	"fmt"
	"reflect"
	"strings"

	dschema "github.com/sky-as-code/nikki-erp/common/dynamicentity/schema"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
)

type EntityValidator struct {
	fields map[string]*dschema.EntityField
}

func NewEntityValidator(schemaFields map[string]*dschema.EntityField) *EntityValidator {
	for _, field := range schemaFields {
		if err := validateEntityFieldName(field); err != nil {
			panic(err)
		}
	}
	return &EntityValidator{
		fields: schemaFields,
	}
}

func (this *EntityValidator) ValidateStruct(entity any, forEdit bool, isQuick bool) ft.ValidationErrors {
	fieldRulesList := make([]*val.FieldRules, 0, len(this.fields))

	for fieldName, field := range this.fields {
		fieldPtr, err := getFieldPointerByTag(entity, fieldName)
		if err != nil {
			continue
		}

		rules := this.buildStructFieldRules(fieldPtr, field, forEdit, isQuick)
		fieldRulesList = append(fieldRulesList, rules)
	}

	return val.ApiBased.ValidateStruct(entity, fieldRulesList...)
}

func (this *EntityValidator) ValidateMap(data map[string]any, forEdit bool, isQuick bool) ft.ValidationErrors {
	return validateMap(data, this.fields, forEdit, isQuick)
}

func (this *EntityValidator) buildStructFieldRules(fieldPtr any, field *dschema.EntityField, forEdit bool, isQuick bool) *val.FieldRules {
	minLength, maxLength := extractLengthOpts(field.Rules())

	modelRules := buildCustomTypeRules(field.DataType(), field.IsRequired() && !forEdit, minLength, maxLength, isQuick)
	if modelRules != nil {
		return val.Field(fieldPtr, modelRules...)
	}
	return val.Field(fieldPtr, buildCoreRules(field, forEdit)...)
}

func validateMap(data map[string]any, fields map[string]*dschema.EntityField, forEdit bool, isQuick bool) ft.ValidationErrors {
	keyRulesList := make([]*val.KeyRules, 0, len(fields))
	targetMap := make(map[string]any)

	for fieldName, field := range fields {
		// Copy from `data` to `targetMap` to set nil to missing key, so that the validator will return the isNil's error
		// instead of the map's "Key missing" error.
		targetMap[fieldName] = data[fieldName]
		rules := buildMapKeyRules(field, forEdit, isQuick)
		keyRulesList = append(keyRulesList, val.Key(fieldName, rules...))
	}

	mapRule := val.Map(keyRulesList...).AllowExtraKeys()
	return val.ApiBased.Validate(targetMap, mapRule)
}

func buildMapKeyRules(field *dschema.EntityField, forEdit bool, isQuick bool) []val.Rule {
	minLength, maxLength := extractLengthOpts(field.Rules())

	modelRules := buildCustomTypeRules(field.DataType(), field.IsRequired() && !forEdit, minLength, maxLength, isQuick)
	if modelRules != nil {
		return modelRules
	}

	coreRules := buildCoreRules(field, forEdit)
	return coreRules
}

func validateEntityFieldName(field *dschema.EntityField) error {
	if strings.TrimSpace(field.Name()) == "" {
		return fmt.Errorf("field name is required")
	}
	return nil
}

func getFieldPointerByTag(entity any, fieldName string) (any, error) {
	rv := reflect.ValueOf(entity)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("entity must be a struct or pointer to struct")
	}

	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		tagVal := field.Tag.Get(dschema.SchemaStructTag)
		if tagVal == "" {
			continue
		}

		dbTagName := strings.Split(tagVal, ",")[0]
		if dbTagName == fieldName {
			fieldValue := rv.Field(i)
			if !fieldValue.CanAddr() {
				return nil, fmt.Errorf("field %s cannot be addressed", fieldName)
			}
			return fieldValue.Addr().Interface(), nil
		}
	}

	return nil, fmt.Errorf("field %s not found with %s tag", fieldName, dschema.SchemaStructTag)
}

func extractLengthOpts(rules []*dschema.FieldRule) (minLength int, maxLength int) {
	for _, rule := range rules {
		ruleName := rule.RuleName()
		if ruleName == dschema.FieldRuleLengthType {
			ruleOptions := rule.RuleOptions()
			lengthArr := ruleOptions.([]int)
			return lengthArr[0], lengthArr[1]
		}
	}
	// Default fallback
	return 1, model.MODEL_RULE_DESC_LENGTH
}

func buildCoreRules(field *dschema.EntityField, forEdit bool) []val.Rule {
	rules := buildIsRequiredRules(field, forEdit)
	rules = append(rules, buildSpecialTypeRules(field.DataType())...)
	rules = append(rules, buildNormalRules(field)...)

	return rules
}

func buildCustomTypeRules(dataType dschema.FieldDataType, isRequired bool, minLength int, maxLength int, isQuick bool) []val.Rule {
	switch dataType {
	case dschema.FieldDataTypeEtag:
		if isQuick {
			return model.EtagRuleQuick(isRequired)
		}
		return model.EtagRule(isRequired)

	case dschema.FieldDataTypeLangJson:
		if isQuick {
			return model.LangJsonRuleQuick(isRequired)
		}
		return model.LangJsonRule(isRequired, minLength, maxLength)

	case dschema.FieldDataTypeLangCode:
		if isQuick {
			return model.LanguageCodeRuleQuick(isRequired)
		}
		return model.LanguageCodeRule(isRequired)

	case dschema.FieldDataTypeModelId, dschema.FieldDataTypeUlid:
		if isQuick {
			return model.IdRuleQuick(isRequired)
		}
		return model.IdRule(isRequired)

	case dschema.FieldDataTypeSlug:
		if isQuick {
			return model.SlugRuleQuick(isRequired)
		}
		return model.SlugRule(isRequired)

	default:
		return nil // Not a special data type
	}
}

func buildIsRequiredRules(field *dschema.EntityField, forEdit bool) []val.Rule {
	rules := make([]val.Rule, 0)

	if field.IsRequired() {
		rules = append(rules,
			val.NotNilWhen(!forEdit),
			val.NotEmptyWhen(!forEdit),
		)
	}

	return rules
}

func buildSpecialTypeRules(dataType dschema.FieldDataType) []val.Rule {
	rules := make([]val.Rule, 0)

	switch dataType {
	case dschema.FieldDataTypeEmail:
		rules = append(rules, val.IsEmail)
	case dschema.FieldDataTypeUrl:
		rules = append(rules, val.IsUrl)
	case dschema.FieldDataTypeUuid:
		rules = append(rules, val.IsUuid)
	}

	return rules
}

func buildNormalRules(field *dschema.EntityField) []val.Rule {
	rules := make([]val.Rule, 0)

	for _, rule := range field.Rules() {
		ruleName := rule.RuleName()
		if ruleName == "" {
			continue
		}

		ruleOptions := rule.RuleOptions()
		switch ruleName {
		case dschema.FieldRuleLengthType:
			lengthArr := ruleOptions.([]int)
			rules = append(rules, val.Length(lengthArr[0], lengthArr[1]))
		case dschema.FieldRuleMaxType:
			rules = append(rules, val.Max(ruleOptions))
		case dschema.FieldRuleMinType:
			rules = append(rules, val.Min(ruleOptions))
		case dschema.FieldRuleOneOfType:
			values := ruleOptions.([]any)
			rules = append(rules, val.OneOf(values...))
		}
	}

	return rules
}

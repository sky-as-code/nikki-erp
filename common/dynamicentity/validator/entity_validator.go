package validator

import (
	"fmt"
	"strings"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicentity/model"
	dschema "github.com/sky-as-code/nikki-erp/common/dynamicentity/schema"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
)

type EntityValidator struct {
	fields  []string
	mapRule *val.MapRule
}

func NewEntityValidator(schemaFields map[string]*dschema.EntityField, forEdit bool) *EntityValidator {
	keyRulesList := make([]*val.KeyRules, 0, len(schemaFields))
	fields := make([]string, 0, len(schemaFields))

	for fieldName, field := range schemaFields {
		if err := validateEntityFieldName(field); err != nil {
			panic(err)
		}
		fields = append(fields, fieldName)
		rules := buildFieldRules(field, forEdit)
		keyRulesList = append(keyRulesList, val.Key(fieldName, rules...))
	}

	mapRule := val.Map(keyRulesList...).AllowExtraKeys()
	return &EntityValidator{
		fields:  fields,
		mapRule: &mapRule,
	}
}

func (this *EntityValidator) ValidateStruct(entity any) ft.ValidationErrors {
	data := dmodel.StructToEntityMap(entity)
	return this.ValidateMap(data)
}

func (this *EntityValidator) ValidateMap(data map[string]any) ft.ValidationErrors {
	targetMap := make(map[string]any)

	for _, fieldName := range this.fields {
		// Copy from `data` to `targetMap` to set nil to missing key, so that the validator will return the isNil's error
		// instead of the map's "Key missing" error.
		targetMap[fieldName] = data[fieldName]
	}

	return val.ApiBased.Validate(targetMap, this.mapRule)
}

func buildFieldRules(fieldDef *dschema.EntityField, forEdit bool) []val.Rule {
	commonRules := buildIsRequiredRules(fieldDef)

	var rules []val.Rule
	minLength, maxLength := extractLengthOpts(fieldDef.Rules())
	rules = buildCustomTypeRules(fieldDef, minLength, maxLength)
	if rules == nil { // Not a custom type
		rules = buildCoreRules(fieldDef)
	}

	return append(commonRules, rules...)
}

func validateEntityFieldName(fieldDef *dschema.EntityField) error {
	if strings.TrimSpace(fieldDef.Name()) == "" {
		return fmt.Errorf("field name is required")
	}
	return nil
}

func extractLengthOpts(rulesDef []*dschema.FieldRule) (minLength int, maxLength int) {
	for _, rule := range rulesDef {
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

func extractArrayLengthRule(rulesDef []*dschema.FieldRule) val.Rule {
	for _, rule := range rulesDef {
		ruleName := rule.RuleName()
		if ruleName == dschema.FieldRuleArrayLengthType {
			ruleOptions := rule.RuleOptions()
			lengthArr := ruleOptions.([]int)
			return val.Length(lengthArr[0], lengthArr[1])
		}
	}
	return nil
}

func buildCoreRules(fieldDef *dschema.EntityField) []val.Rule {
	rules := buildSpecialTypeRules(fieldDef.DataType())
	rules = append(rules, buildNormalRules(fieldDef)...)

	if fieldDef.IsArray() {
		rules = buildArrayRules(fieldDef, rules)
	}

	return rules
}

func buildIsRequiredRules(fieldDef *dschema.EntityField) []val.Rule {
	rules := make([]val.Rule, 0)

	if fieldDef.IsRequired() {
		rules = append(rules,
			val.NotNil,
			val.NotEmpty,
		)
	}

	return rules
}

func buildCustomTypeRules(fieldDef *dschema.EntityField, minLength int, maxLength int) []val.Rule {
	dataType := fieldDef.DataType()

	var rules []val.Rule
	switch dataType {
	case dschema.FieldDataTypeEtag:
		rules = model.EtagRules()

	case dschema.FieldDataTypeLangJson:
		rules = model.LangJsonRules(minLength, maxLength)

	case dschema.FieldDataTypeLangCode:
		rules = model.LanguageCodeRules()

	case dschema.FieldDataTypeModelId, dschema.FieldDataTypeUlid:
		rules = model.IdRules()

	case dschema.FieldDataTypeSlug:
		rules = model.SlugRules()

	default:
		rules = nil // Not a special data type
	}

	if fieldDef.IsArray() {
		return buildArrayRules(fieldDef, rules)
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

func buildNormalRules(fieldDef *dschema.EntityField) []val.Rule {
	rules := make([]val.Rule, 0)

	for _, rule := range fieldDef.Rules() {
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

func buildArrayRules(fieldDef *dschema.EntityField, itemRules []val.Rule) []val.Rule {
	arrayRules := make([]val.Rule, 0)

	arrayLengthRule := extractArrayLengthRule(fieldDef.Rules())
	if arrayLengthRule != nil {
		arrayRules = append(arrayRules, arrayLengthRule)
	}
	itemRules = append(itemRules, val.NotNil, val.NotEmpty)
	return append(arrayRules, val.Each(itemRules...))
}

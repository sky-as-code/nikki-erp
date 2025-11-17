package validator

import (
	dschema "github.com/sky-as-code/nikki-erp/common/dynamicentity/schema"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	val "github.com/sky-as-code/nikki-erp/common/validator"
)

type AdhocValidator struct {
	schema *dschema.AdhocSchema
}

func NewAdhocValidator(schema *dschema.AdhocSchema) *AdhocValidator {
	validateAdhocFieldNames(schema.Fields())

	return &AdhocValidator{
		schema: schema,
	}
}

func validateAdhocFieldNames(fields map[string]*dschema.AdhocField) {
	for _, field := range fields {
		if field.IsHolder() {
			validateAdhocFieldNames(field.HolderSchema().Fields())
			continue
		}
		if err := validateEntityFieldName(field.Field()); err != nil {
			panic(err)
		}
	}
}

func (this *AdhocValidator) ValidateMap(data map[string]any, forEdit bool, isQuick bool) ft.ValidationErrors {
	mapRule, targetMap := this.buildMapRules(data, this.schema, forEdit, isQuick)
	return val.ApiBased.Validate(targetMap, mapRule)
}

func (this *AdhocValidator) buildMapRules(
	data map[string]any,
	schema *dschema.AdhocSchema,
	forEdit bool,
	isQuick bool,
) (mapRule val.MapRule, newData map[string]any) {
	fields := schema.Fields()
	keyRulesList := make([]*val.KeyRules, 0, len(fields))
	newData = make(map[string]any)

	for fieldName, field := range fields {
		// Copy from `data` to `targetMap` to set nil to missing key, so that the validator will return the isNil's error
		// instead of the map's "Key missing" error.
		newData[fieldName] = data[fieldName]
		rules := buildIsRequiredRules(field.Field(), forEdit)

		if data[fieldName] != nil {
			if field.IsHolder() {
				// rules = append(rules, buildIsRequiredRules(field.Field(), forEdit)...)
				subMap, ok := data[fieldName].(map[string]any)

				// Generate other rules for non-nil data.
				// Only IsRequired is applicable to nil field.
				if ok {
					var subMapRule val.MapRule
					subMapRule, subMap = this.buildMapRules(subMap, field.HolderSchema(), forEdit, isQuick)
					newData[fieldName] = subMap
					rules = append(rules, subMapRule)
				}
			} else {
				rules = this.builLeaveKeyRules(field, forEdit, isQuick)
			}
		}
		keyRulesList = append(keyRulesList, val.Key(fieldName, rules...))
	}

	mapRule = val.Map(keyRulesList...).AllowExtraKeys()
	return
}

func (this *AdhocValidator) builLeaveKeyRules(adhocF *dschema.AdhocField, forEdit bool, isQuick bool) []val.Rule {
	minLength, maxLength := extractLengthOpts(adhocF.Field().Rules())

	modelRules := buildCustomTypeRules(adhocF.Field().DataType(), adhocF.IsRequired() && !forEdit, minLength, maxLength, isQuick)
	if modelRules != nil {
		return modelRules
	}

	coreRules := buildCoreRules(adhocF.Field(), forEdit)
	return coreRules
}

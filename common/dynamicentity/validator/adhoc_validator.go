package validator

import (
	invopop "github.com/invopop/validation"
	"go.bryk.io/pkg/errors"

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
	vErrs := this.validateMapRecursive(data, this.schema, forEdit, isQuick)
	return vErrs
}

func (this *AdhocValidator) validateMapRecursive(
	data map[string]any,
	schema *dschema.AdhocSchema,
	forEdit bool,
	isQuick bool,
) ft.ValidationErrors {
	fieldMap := schema.Fields()
	keyRules := make([]*val.KeyRules, 0, len(fieldMap))
	dataAllKeys := make(map[string]any)
	vErrs := ft.NewValidationErrors()

	for fieldName, fieldDef := range fieldMap {
		// Copy from `data` to `dataAllKeys` to set nil to missing key, so that the validator will return the isNil's error
		// instead of the map's "Key missing" error.
		dataAllKeys[fieldName] = data[fieldName]
		rules := buildIsRequiredRules(fieldDef.Field(), forEdit)

		if dataAllKeys[fieldName] != nil {
			if fieldDef.IsHolder() {
				subMap, ok := dataAllKeys[fieldName].(map[string]any)

				if !ok {
					subMap = make(map[string]any)
				}
				subVErrs := this.validateMapRecursive(subMap, fieldDef.HolderSchema(), forEdit, isQuick)
				vErrs.Merge(subVErrs, fieldName+"$")
			} else {
				rules = buildFieldRules(fieldDef.Field(), forEdit, isQuick)
			}
		}
		keyRules = append(keyRules, val.Key(fieldName, rules...))
	}

	rawErr := val.Map(keyRules...).AllowExtraKeys().Validate(dataAllKeys)
	if rawErr == nil {
		return vErrs
	}
	invopopErr, isOk := rawErr.(invopop.Errors)
	if isOk {
		vErrs.Merge(ft.NewValidationErrorsFromInvopop(invopopErr))
	} else {
		panic(errors.Wrap(rawErr, "failed to validate struct"))
	}
	return vErrs
}

// func (this *AdhocValidator) builLeaveKeyRules(fieldDef *dschema.AdhocField, forEdit bool, isQuick bool) []val.Rule {
// 	minLength, maxLength := extractLengthOpts(fieldDef.Field().Rules())

// 	modelRules := buildCustomTypeRules(fieldDef.Field(), fieldDef.IsRequired() && !forEdit, minLength, maxLength, isQuick)
// 	if modelRules != nil {
// 		return modelRules
// 	}

// 	coreRules := buildCoreRules(fieldDef.Field(), forEdit, isQuick)
// 	return coreRules
// }

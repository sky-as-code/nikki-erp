package model

import (
	"reflect"

	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
)

// RequiredFields returns, per value of this enum field, the fields that must have a value while
// this field holds it. The returned map is the field's own, so callers must not mutate it.
func (this *ModelField) RequiredFields() map[string][]string {
	return this.requiredFields
}

// OptionalFields returns, per value of this enum field, the fields that are relevant while this
// field holds it. The engine never enforces them; they only tell a client what to show.
func (this *ModelField) OptionalFields() map[string][]string {
	return this.optionalFields
}

// HasConditionalRequiredFields tells whether any field declares required_fields, so a caller
// can skip loading the stored row it would otherwise need for the update-time check.
func (this *ModelSchema) HasConditionalRequiredFields() bool {
	for _, field := range this.fields {
		if len(field.requiredFields) > 0 {
			return true
		}
	}
	return false
}

// ValidateConditionalFields enforces required_fields against the effective record: a key present
// in input wins, otherwise the stored value counts. Pass stored = nil on create. Fields already
// failing in input validation must not reach here; the update flow runs it after Validate.
func (this *ModelSchema) ValidateConditionalFields(input DynamicFields, stored DynamicFields) ft.ClientErrors {
	var errs ft.ClientErrors
	this.appendConditionalFieldErrors(&errs, input, stored)
	return errs
}

// appendConditionalFieldErrors skips a source whose own value already failed validation, and a
// listed field that already carries an error, so one bad input never reports twice.
func (this *ModelSchema) appendConditionalFieldErrors(errs *ft.ClientErrors, input, stored DynamicFields) {
	for _, name := range this.fieldsOrder {
		field := this.fields[name]
		if len(field.requiredFields) == 0 || errs.Has(name) {
			continue
		}
		key, ok := enumKeyOf(effectiveValue(name, input, stored))
		if !ok {
			continue
		}
		for _, listed := range field.requiredFields[key] {
			if errs.Has(listed) || hasValue(effectiveValue(listed, input, stored)) {
				continue
			}
			errs.Append(*NewMissingFieldErr(listed))
		}
	}
}

func effectiveValue(name string, input, stored DynamicFields) any {
	if val, ok := input[name]; ok {
		return val
	}
	return stored[name]
}

func enumKeyOf(val any) (string, bool) {
	rv := derefValue(val)
	if !rv.IsValid() || rv.Kind() != reflect.String || rv.Len() == 0 {
		return "", false
	}
	return rv.String(), true
}

// hasValue differs from isNilOrEmpty on purpose: a numeric 0 or a false is a value the user chose,
// and a required field holding it is satisfied.
func hasValue(val any) bool {
	rv := derefValue(val)
	if !rv.IsValid() {
		return false
	}
	switch rv.Kind() {
	case reflect.String, reflect.Slice, reflect.Array, reflect.Map:
		return rv.Len() > 0
	}
	return true
}

func derefValue(val any) reflect.Value {
	rv := reflect.ValueOf(val)
	for rv.IsValid() && (rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface) {
		if rv.IsNil() {
			return reflect.Value{}
		}
		rv = rv.Elem()
	}
	return rv
}

func cloneConditionalFields(source map[string][]string) map[string][]string {
	if source == nil {
		return nil
	}
	cloned := make(map[string][]string, len(source))
	for key, names := range source {
		cloned[key] = append([]string(nil), names...)
	}
	return cloned
}

// validateConditionalFields checks required_fields / optional_fields at Build(). The source must
// be a scalar enum_string: JSON map keys are strings, so an enum_int32 key could never match.
func validateConditionalFields(schema *ModelSchema) error {
	for _, name := range schema.fieldsOrder {
		field := schema.fields[name]
		if len(field.requiredFields) == 0 && len(field.optionalFields) == 0 {
			continue
		}
		if err := assertConditionalSource(field); err != nil {
			return err
		}
		allowed := getEnumStringValues(field.dataType.Options())
		if err := assertConditionalMap(schema, field, "required_fields", field.requiredFields, allowed); err != nil {
			return err
		}
		if err := assertConditionalMap(schema, field, "optional_fields", field.optionalFields, allowed); err != nil {
			return err
		}
		if err := assertConditionalMapsDisjoint(field); err != nil {
			return err
		}
	}
	return nil
}

func assertConditionalSource(field *ModelField) error {
	if field.dataType == nil || field.dataType.String() != FieldDataTypeNameEnumString || field.dataType.IsArray() {
		return errors.Errorf(
			"validateConditionalFields: field '%s': required_fields/optional_fields need a scalar enum_string source",
			field.name)
	}
	return nil
}

func assertConditionalMap(
	schema *ModelSchema, field *ModelField, option string, conditional map[string][]string, allowed []string,
) error {
	for key, names := range conditional {
		if !containsString(allowed, key) {
			return errors.Errorf("validateConditionalFields: field '%s': %s key '%s' is not one of its enum values",
				field.name, option, key)
		}
		for _, listed := range names {
			if err := assertConditionalTarget(schema, field, option, listed); err != nil {
				return err
			}
		}
	}
	return nil
}

func assertConditionalTarget(schema *ModelSchema, field *ModelField, option string, listed string) error {
	if listed == field.name {
		return errors.Errorf("validateConditionalFields: field '%s': %s cannot list the field itself",
			field.name, option)
	}
	target, ok := schema.Field(listed)
	if !ok {
		return errors.Errorf("validateConditionalFields: field '%s': %s lists undefined field '%s'",
			field.name, option, listed)
	}
	// A virtual field is never written, so requiring it would reject every write.
	if option == "required_fields" && !target.IsPersisted() {
		return errors.Errorf("validateConditionalFields: field '%s': %s lists virtual field '%s'",
			field.name, option, listed)
	}
	return nil
}

func assertConditionalMapsDisjoint(field *ModelField) error {
	for key, requiredNames := range field.requiredFields {
		for _, listed := range field.optionalFields[key] {
			if containsString(requiredNames, listed) {
				return errors.Errorf(
					"validateConditionalFields: field '%s': '%s' is both required and optional for '%s'",
					field.name, listed, key)
			}
		}
	}
	return nil
}

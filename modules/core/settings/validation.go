package settings

import (
	"fmt"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
)

// ValidateItems checks every item against the schema the module registered for the level: the
// name must be declared and the value must fit its field. It is the schema half of a module's
// validation; the Settings module runs the same check again when it persists.
func ValidateItems(schema *dmodel.ModelSchema, items []SetSettingItem) ft.ClientErrors {
	vErrs := ft.ClientErrors{}
	for _, item := range items {
		field, ok := schema.Field(item.Name)
		if !ok {
			vErrs.Append(*UnknownSettingError(item.Name))
			continue
		}
		if vErr := ValidateValue(field, item.Value); vErr != nil {
			vErrs.Append(*vErr)
		}
	}
	return vErrs
}

// UnknownSettingError is the caller's error for a name no schema declares, not a server fault.
func UnknownSettingError(name string) *ft.ClientErrorItem {
	return ft.NewValidationError(name, "settings.unknown_setting", "is not a setting of this module at this level")
}

// ValidateValue checks one value against its declared field. A nil value clears the setting back
// to its declared default, which every field accepts.
func ValidateValue(field *dmodel.ModelField, value any) *ft.ClientErrorItem {
	if value == nil {
		return nil
	}
	if _, vErr := field.Validate(value, true); vErr != nil {
		return vErr
	}
	return assertWithinDeclaredRange(field, value)
}

// assertWithinDeclaredRange rejects a numeric value outside its field's declared bounds.
//
// This duplicates a check ModelField.Validate is supposed to perform, and does so deliberately.
// Validate treats a numeric ZERO as an absent value platform-wide — isNilOrEmpty returns true for
// it — so it returns before the range check and a submitted 0 is stored as a real value. For an
// ordinary field that is harmless, but a settings value arrives straight from an API caller and is
// read back as policy: a session timeout of 0 means every session in the tenant expires
// immediately, locking everyone out through the interface that set it.
//
// Scoped to numbers, because zero is the only value the platform helper misreads. Fixing this in
// isNilOrEmpty would change validation for every numeric field in the product, which is a larger
// decision than this module should make on its own.
func assertWithinDeclaredRange(field *dmodel.ModelField, raw any) *ft.ClientErrorItem {
	number, isNumber := asFloat(raw)
	if !isNumber {
		return nil
	}

	limits, ok := declaredNumericRange(field)
	if !ok {
		return nil
	}
	if number < limits[0] || number > limits[1] {
		return ft.NewValidationError(field.Name(), "settings.value_out_of_range",
			fmt.Sprintf("must be between %v and %v", limits[0], limits[1]))
	}
	return nil
}

// asFloat reports whether a decoded JSON value is a number, in whatever concrete type it arrived
// as. A JSON body decodes numbers as float64, but a value round-tripped through the model layer
// can be an int32, so both are accepted.
func asFloat(raw any) (float64, bool) {
	switch number := raw.(type) {
	case float64:
		return number, true
	case float32:
		return float64(number), true
	case int:
		return float64(number), true
	case int32:
		return float64(number), true
	case int64:
		return float64(number), true
	}
	return 0, false
}

// declaredNumericRange reads a field's declared bounds, whatever slice type the data type
// constructor stored them as.
func declaredNumericRange(field *dmodel.ModelField) ([2]float64, bool) {
	raw, ok := field.DataType().Options()[dmodel.FieldDataTypeOptRange]
	if !ok {
		return [2]float64{}, false
	}

	switch limits := raw.(type) {
	case []int32:
		if len(limits) == 2 {
			return [2]float64{float64(limits[0]), float64(limits[1])}, true
		}
	case []int64:
		if len(limits) == 2 {
			return [2]float64{float64(limits[0]), float64(limits[1])}, true
		}
	case []int:
		if len(limits) == 2 {
			return [2]float64{float64(limits[0]), float64(limits[1])}, true
		}
	}
	return [2]float64{}, false
}

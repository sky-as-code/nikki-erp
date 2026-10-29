package model

import (
	"regexp"
	"strings"

	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
)

// resourceCodeRegex is shared verbatim with the frontend's nikkiResourceCode input mask.
var resourceCodeRegex = regexp.MustCompile(`^[A-Z0-9-]+$`)

// FieldDataTypeResourceCode is the identifier code a user types for a resource (warehouse, tax, channel...).
// Values are never case-converted: a lowercase code is rejected rather than silently rewritten.
func FieldDataTypeResourceCode(minLength int, maxLength int) FieldDataType {
	validateRangeOrPanic(minLength, maxLength, "FieldDataTypeResourceCode")
	return fieldDataTypeResourceCode{fieldDataTypeBase{
		name:    FieldDataTypeNameResourceCode,
		options: FieldDataTypeOptions{FieldDataTypeOptLength: []int{minLength, maxLength}},
	}}
}

func IsValidResourceCode(code string) bool {
	return resourceCodeRegex.MatchString(code)
}

// NormalizeResourceCode maps a legacy spelling ("vdmc", "VN_VAT_10") onto the strict form, so callers
// configured before the codes were renamed still resolve. It is for lookups and service-side input
// only: the data type itself never rewrites what a user typed.
func NormalizeResourceCode(code string) string {
	return strings.ReplaceAll(strings.ToUpper(strings.TrimSpace(code)), "_", "-")
}

func NewInvalidResourceCodeErr(field string) *ft.ClientErrorItem {
	return ft.NewValidationError(field, ft.ErrorKey("err_invalid_resource_code"),
		"Only uppercase letters, digits and hyphens are allowed")
}

type fieldDataTypeResourceCode struct{ fieldDataTypeBase }

func (this fieldDataTypeResourceCode) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeResourceCode) Validate(val value) (value, *ft.ClientErrorItem) {
	if this.isArray {
		return validateArrayAfterTryConvert(this, val, this.validateScalar)
	}
	return validateScalarAfterTryConvert(this, val, this.validateScalar)
}

func (this fieldDataTypeResourceCode) validateScalar(val value) (value, *ft.ClientErrorItem) {
	sanitized, clientErr := validateStringBase(val, this.options)
	if clientErr != nil {
		return Value(nil), clientErr
	}
	// An empty code passed the length check, so it is allowed: only an optional field (min 0) gets here.
	if text := (*sanitized.Get()).(string); text != "" && !IsValidResourceCode(text) {
		return Value(nil), NewInvalidResourceCodeErr("")
	}
	return sanitized, nil
}

func (this fieldDataTypeResourceCode) TryConvert(val any, _ FieldDataTypeOptions) (value, error) {
	if this.isArray {
		return tryConvertResourceCodeArrayValue(val)
	}
	text, err := toString(val)
	if err != nil {
		return Value(nil), err
	}
	return Value(strings.TrimSpace(text)), nil
}

func tryConvertResourceCodeArrayValue(val any) (value, error) {
	stringsValue, err := tryConvertStringArrayValue(val)
	if err != nil {
		return Value(nil), err
	}
	if stringsValue.Get() == nil {
		return Value(nil), errors.New("tryConvertResourceCodeArrayValue: value cannot be nil")
	}
	raw := (*stringsValue.Get()).([]string)
	out := make([]string, len(raw))
	for i := range raw {
		out[i] = strings.TrimSpace(raw[i])
	}
	return Value(out), nil
}

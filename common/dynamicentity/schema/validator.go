package schema

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"text/template"

	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
)

func ToClientErrorItem(err *ValidationErrorItem) ft.ClientErrorItem {
	if err == nil {
		return ft.ClientErrorItem{}
	}
	return ft.ClientErrorItem{
		Field:   err.Field,
		Key:     err.Key,
		Message: err.Message,
		Type:    ft.ClientErrorTypeValidation,
		Vars:    err.Vars,
	}
}

type ValidationErrors map[string]ValidationErrorItem

func (this ValidationErrors) AddItem(item ValidationErrorItem) {
	this[item.Field] = item
}

// ValidationErrorItem implements the error interface with code, message template, and vars for substitution.
type ValidationErrorItem struct {
	Field   string // field name that has the violation, set by EntityField.Validate()
	Key     string
	Message string
	Vars    map[string]any
}

// Error returns the message with variables substituted. Implements the error interface.
func (this *ValidationErrorItem) Error() string {
	return this.String()
}

// String returns the message with variables substituted.
func (this *ValidationErrorItem) String() string {
	if this == nil || this.Message == "" {
		return ""
	}
	if len(this.Vars) == 0 {
		return this.Message
	}
	tmpl, err := template.New("validation").Parse(this.Message)
	if err != nil {
		return this.Message
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, this.Vars); err != nil {
		return this.Message
	}
	return buf.String()
}

func (this *ValidationErrorItem) ToClientErrorItem() *ft.ClientErrorItem {
	return &ft.ClientErrorItem{
		Field:   this.Field,
		Key:     this.Key,
		Message: this.Message,
		Type:    ft.ClientErrorTypeValidation,
		Vars:    this.Vars,
	}
}

// ValidateMax validates that value is not greater than max. Supports numbers, strings (length), slices (length).
func ValidateMax(value any, opts any) *ft.ClientErrorItem {
	maxVal, err := toComparableThreshold(opts)
	if err != nil {
		panic(errors.Wrap(err, "invalid max option"))
	}
	val, err := toComparableValue(value)
	if err != nil {
		return ft.NewAnonymousValidationError("common.err_invalid_value_type_number", "invalid value type, must be a number", nil)
	}
	if !isLessOrEqual(val, maxVal) {
		return ft.NewAnonymousValidationError("common.err_greater_than_max", "must not be greater than {{.max}}", map[string]any{"max": fmt.Sprint(opts)})
	}
	return nil
}

// ValidateMin validates that value is not less than min. Supports numbers, strings (length), slices (length).
func ValidateMin(value any, opts any) *ft.ClientErrorItem {
	minVal, err := toComparableThreshold(opts)
	if err != nil {
		panic(errors.Wrap(err, "invalid min option"))
	}
	val, err := toComparableValue(value)
	if err != nil {
		return ft.NewAnonymousValidationError("common.err_invalid_value_type_number", "invalid value type, must be a number", nil)
	}
	if !isGreaterOrEqual(val, minVal) {
		return ft.NewAnonymousValidationError("common.err_less_than_min", "must not be less than {{.min}}", map[string]any{"min": fmt.Sprint(opts)})
	}
	return nil
}

// ValidateLength validates that string or slice length is between min and max (inclusive).
// opts must be []int{min, max}.
func ValidateLength(value any, opts any) *ft.ClientErrorItem {
	arr, ok := opts.([]int)
	if !ok || len(arr) < 2 {
		panic(errors.New("invalid length option: must be []int{min, max} with at least 2 elements"))
	}
	minLen, maxLen := arr[0], arr[1]
	length, err := getLength(value)
	if err != nil {
		return ft.NewAnonymousValidationError("common.err_invalid_value_type", "invalid value type", nil)
	}
	if length < minLen || length > maxLen {
		return ft.NewAnonymousValidationError("common.err_length_out_of_range", "must have length between {{.min}} and {{.max}}", map[string]any{"min": minLen, "max": maxLen})
	}
	return nil
}

// ValidateArrayLength validates that slice/array length is between min and max (inclusive).
// opts must be []int{min, max}.
func ValidateArrayLength(value any, opts any) *ft.ClientErrorItem {
	arr, ok := opts.([]int)
	if !ok || len(arr) < 2 {
		panic(errors.New("invalid arrlength option: must be []int{min, max} with at least 2 elements"))
	}
	minLen, maxLen := arr[0], arr[1]
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return ft.NewAnonymousValidationError("common.err_invalid_value_type_array", "invalid value type, must be an array", nil)
	}
	n := rv.Len()
	if n < minLen || n > maxLen {
		return ft.NewAnonymousValidationError("common.err_array_length_out_of_range", "array length must be between {{.min}} and {{.max}}", map[string]any{"min": minLen, "max": maxLen})
	}
	return nil
}

// ValidateOneOf validates that value is one of the allowed values.
// opts must be []any of allowed values.
func ValidateOneOf(value any, opts any) *ft.ClientErrorItem {
	values, ok := opts.([]any)
	if !ok || len(values) == 0 {
		panic(errors.New("invalid oneOf option: must be non-empty []any"))
	}
	valRef := reflect.ValueOf(value)
	for _, allowed := range values {
		allowedRef := reflect.ValueOf(allowed)
		if reflect.DeepEqual(valRef.Interface(), allowedRef.Interface()) {
			return nil
		}
	}
	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = fmt.Sprint(v)
	}
	return ft.NewAnonymousValidationError("common.err_not_one_of", "must be one of: {{.allowed}}", map[string]any{"allowed": strings.Join(parts, ", ")})
}

// --- Data type validation helpers ---

func ValidateNotEmpty(value any) *ft.ClientErrorItem {
	if isEmptyValue(value) {
		return ft.NewAnonymousValidationError("common.err_not_empty_required", "must not be empty", nil)
	}
	return nil
}

func ValidateNotNil(value any) *ft.ClientErrorItem {
	if value == nil {
		return ft.NewAnonymousValidationError("common.err_not_nil_required", "must not be nil", nil)
	}
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return ft.NewAnonymousValidationError("common.err_not_nil_required", "must not be nil", nil)
	}
	return nil
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func errIncompatibleDataType() *ft.ClientErrorItem {
	return ft.NewAnonymousValidationError("common.err_incompatible_data_type", "incompatible data type", nil)
}

func ValidateEmail(value string) *ft.ClientErrorItem {
	if !emailRegex.MatchString(value) {
		return ft.NewAnonymousValidationError("common.err_invalid_email", "must be a valid email address", nil)
	}
	return nil
}

var urlRegex = regexp.MustCompile(`^https?://[^\s]+$`)

func ValidateUrl(value string) *ft.ClientErrorItem {
	if !urlRegex.MatchString(value) {
		return ft.NewAnonymousValidationError("common.err_invalid_url", "must be a valid URL", nil)
	}
	return nil
}

var uuidRegex = regexp.MustCompile(
	`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)

func ValidateUuid(value string) *ft.ClientErrorItem {
	if !uuidRegex.MatchString(value) {
		return ft.NewAnonymousValidationError("common.err_invalid_uuid", "must be a valid UUID", nil)
	}
	return nil
}

func ValidatePattern(value string, re *regexp.Regexp) *ft.ClientErrorItem {
	if re == nil {
		return nil
	}
	if !re.MatchString(value) {
		return ft.NewAnonymousValidationError("common.err_format_mismatch", "must match the required format", nil)
	}
	return nil
}

// --- Comparable value helpers ---

type comparableKind int

const (
	kindNumber comparableKind = iota
	kindLength
)

func toComparableThreshold(opts any) (float64, error) {
	if opts == nil {
		return 0, errors.New("threshold cannot be nil")
	}
	switch v := opts.(type) {
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	default:
		return 0, errors.Errorf("unsupported threshold type %T", opts)
	}
}

func toComparableValue(value any) (float64, error) {
	if value == nil {
		return 0, errors.New("value cannot be nil")
	}
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(rv.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(rv.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return rv.Float(), nil
	default:
		return 0, errors.Errorf("cannot compare type %T for max/min", value)
	}
}

func isLessOrEqual(val float64, threshold float64) bool {
	return val <= threshold
}

func isGreaterOrEqual(val float64, threshold float64) bool {
	return val >= threshold
}

func getLength(value any) (int, error) {
	if value == nil {
		return 0, errors.New("value cannot be nil")
	}
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.String:
		return rv.Len(), nil
	case reflect.Slice, reflect.Map, reflect.Array:
		return rv.Len(), nil
	default:
		return 0, errors.Errorf("length rule applies to strings and slices, got %T", value)
	}
}

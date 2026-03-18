package schema

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.bryk.io/pkg/errors"
	"go.bryk.io/pkg/ulid"

	"github.com/sky-as-code/nikki-erp/common/defense"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
)

// --- Factory functions (replace package-level vars) ---
// Scalar types are created by default. Use .ArrayType() to get array variant.

func FieldDataTypeEmail() FieldDataType {
	return fieldDataTypeEmail{fieldDataTypeBase{name: "email", options: nil}}
}

func FieldDataTypePhone() FieldDataType {
	return fieldDataTypePhone{fieldDataTypeBase{name: "phone", options: nil}}
}

func FieldDataTypeString(sanitizeType ...SanitizeType) FieldDataType {
	st := SanitizeTypePlainText
	if len(sanitizeType) > 0 && sanitizeType[0] != "" {
		st = sanitizeType[0]
	}
	opts := FieldDataTypeOptions{FieldDataTypeOptSanitizeType: st}
	return fieldDataTypeString{fieldDataTypeBase{name: "string", options: opts}}
}

func FieldDataTypeSecret() FieldDataType {
	return fieldDataTypeSecret{fieldDataTypeBase{name: "secret", options: nil}}
}

func FieldDataTypeUrl() FieldDataType {
	return fieldDataTypeUrl{fieldDataTypeBase{name: "url", options: nil}}
}

func FieldDataTypeUlid() FieldDataType {
	return fieldDataTypeUlid{fieldDataTypeBase{name: "ulid", options: nil}}
}

func FieldDataTypeUuid() FieldDataType {
	return fieldDataTypeUuid{fieldDataTypeBase{name: "uuid", options: nil}}
}

func FieldDataTypeInteger() FieldDataType {
	return fieldDataTypeInteger{fieldDataTypeBase{name: "integer", options: nil}}
}

func FieldDataTypeFloat(precision int) FieldDataType {
	opts := FieldDataTypeOptions{FieldDataTypeOptPrecision: precision}
	return fieldDataTypeFloat{fieldDataTypeBase{name: "float", options: opts}}
}

func FieldDataTypeBoolean() FieldDataType {
	return fieldDataTypeBoolean{fieldDataTypeBase{name: "boolean", options: nil}}
}

func FieldDataTypeDate() FieldDataType {
	return fieldDataTypeDate{fieldDataTypeBase{name: "date", options: nil}}
}

func FieldDataTypeTime() FieldDataType {
	return fieldDataTypeTime{fieldDataTypeBase{name: "time", options: nil}}
}

func FieldDataTypeDateTime() FieldDataType {
	return fieldDataTypeDateTime{fieldDataTypeBase{name: "dateTime", options: nil}}
}

func FieldDataTypeEnumString(enumValues []string) FieldDataType {
	opts := FieldDataTypeOptions{FieldDataTypeOptEnumValues: enumValues}
	return fieldDataTypeEnumString{fieldDataTypeBase{name: "enumString", options: opts}}
}

func FieldDataTypeEnumNumber(enumValues []int64) FieldDataType {
	opts := FieldDataTypeOptions{FieldDataTypeOptEnumValues: enumValues}
	return fieldDataTypeEnumNumber{fieldDataTypeBase{name: "enumNumber", options: opts}}
}

func FieldDataTypeEtag() FieldDataType {
	return fieldDataTypeEtag{fieldDataTypeBase{name: "nikkiEtag", options: nil}}
}

func FieldDataTypeLangJson(sanitizeType ...SanitizeType) FieldDataType {
	st := SanitizeTypePlainText
	if len(sanitizeType) > 0 && sanitizeType[0] != "" {
		st = sanitizeType[0]
	}
	opts := FieldDataTypeOptions{FieldDataTypeOptSanitizeType: st}
	return fieldDataTypeLangJson{fieldDataTypeBase{name: "nikkiLangJson", options: opts}}
}

func FieldDataTypeLangCode() FieldDataType {
	return fieldDataTypeLangCode{fieldDataTypeBase{name: "nikkiLangCode", options: nil}}
}

func FieldDataTypeModelId() FieldDataType {
	return fieldDataTypeModelId{fieldDataTypeBase{name: "nikkiModelId", options: nil}}
}

func FieldDataTypeSlug() FieldDataType {
	return fieldDataTypeSlug{fieldDataTypeBase{name: "nikkiSlug", options: nil}}
}

// FieldDataTypeEntity represents a virtual/implicit field that holds a related entity or slice of entities.
// It is not persisted as a DB column; it is used for graph traversal and API response expansion.
func FieldDataTypeEntity() FieldDataType {
	return fieldDataTypeEntity{fieldDataTypeBase{name: "entity", options: nil}}
}

type fieldDataTypeEntity struct{ fieldDataTypeBase }

func (this fieldDataTypeEntity) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeEntity) Validate(value any) (any, *ft.ClientErrorItem) {
	return value, nil
}

func (this fieldDataTypeEntity) TryConvert(value any, _ FieldDataTypeOptions) (any, error) {
	return value, nil
}

// FieldDataType defines the interface for dynamic entity field data types.
// Validate returns (validatedValue, nil) on success or (nil, ValidationError) on failure.
// Options are embedded in the data type; Validate uses them internally.
type FieldDataType interface {
	Validate(value any) (any, *ft.ClientErrorItem)
	String() string
	ArrayType() FieldDataType
	IsArray() bool
	Options() FieldDataTypeOptions
	TryConvert(value any, options FieldDataTypeOptions) (any, error)
}

// fieldDataTypeBase provides common behavior for simple string-based types.
type fieldDataTypeBase struct {
	name    string
	isArray bool
	options FieldDataTypeOptions
}

func (this fieldDataTypeBase) String() string {
	return this.name
}

func (this fieldDataTypeBase) IsArray() bool {
	return this.isArray
}

func (this fieldDataTypeBase) Options() FieldDataTypeOptions {
	return this.options
}

// --- String-like types ---

type fieldDataTypeEmail struct{ fieldDataTypeBase }

func (this fieldDataTypeEmail) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeEmail) Validate(value any) (any, *ft.ClientErrorItem) {
	if err := ValidateNotEmpty(value); err != nil {
		return nil, err
	}
	if err := ValidateEmail(value); err != nil {
		return nil, err
	}
	return value, nil
}

func (this fieldDataTypeEmail) TryConvert(value any, _ FieldDataTypeOptions) (any, error) {
	return toString(value)
}

type fieldDataTypePhone struct{ fieldDataTypeBase }

func (this fieldDataTypePhone) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypePhone) Validate(value any) (any, *ft.ClientErrorItem) {
	s, err := toString(value)
	if err != nil {
		return nil, errIncompatibleDataType()
	}
	if strings.TrimSpace(s) == "" {
		return nil, &ft.ClientErrorItem{Key: "phone_empty", Message: "phone cannot be empty", Vars: nil}
	}
	return value, nil
}

func (this fieldDataTypePhone) TryConvert(value any, _ FieldDataTypeOptions) (any, error) {
	return toString(value)
}

type fieldDataTypeString struct{ fieldDataTypeBase }

func (this fieldDataTypeString) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeString) Validate(value any) (any, *ft.ClientErrorItem) {
	_, err := toString(value)
	if err != nil {
		return nil, errIncompatibleDataType()
	}
	return sanitizeStringValue(value, this.options)
}

func sanitizeStringValue(value any, options FieldDataTypeOptions) (any, *ft.ClientErrorItem) {
	if options == nil {
		return value, nil
	}
	st := options[FieldDataTypeOptSanitizeType].(SanitizeType)
	if rv := reflect.ValueOf(value); rv.Kind() == reflect.Slice {
		return sanitizeStringSlice(value, st)
	}
	return sanitizeStringScalar(value, st)
}

func sanitizeStringScalar(value any, st SanitizeType) (any, *ft.ClientErrorItem) {
	switch v := value.(type) {
	case string:
		return sanitizeByType(v, st), nil
	case *string:
		if v == nil {
			return value, nil
		}
		return util.ToPtr(sanitizeByType(*v, st)), nil
	default:
		return value, nil
	}
}

func sanitizeStringSlice(value any, st SanitizeType) (any, *ft.ClientErrorItem) {
	rv := reflect.ValueOf(value)
	n := rv.Len()
	result := make([]any, n)
	for i := 0; i < n; i++ {
		sanitized, err := sanitizeStringScalar(rv.Index(i).Interface(), st)
		if err != nil {
			return nil, err
		}
		result[i] = sanitized
	}
	return result, nil
}

func sanitizeByType(s string, t SanitizeType) string {
	switch t {
	case SanitizeTypeHtml:
		return defense.SanitizeRichText(s)
	case SanitizeTypePlainText:
		return defense.SanitizePlainText(s, true)
	default:
		return s
	}
}

func (this fieldDataTypeString) TryConvert(value any, _ FieldDataTypeOptions) (any, error) {
	return toString(value)
}

type fieldDataTypeSecret struct{ fieldDataTypeBase }

func (this fieldDataTypeSecret) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeSecret) Validate(value any) (any, *ft.ClientErrorItem) {
	_, err := toString(value)
	if err != nil {
		return nil, errIncompatibleDataType()
	}
	return value, nil
}

func (this fieldDataTypeSecret) TryConvert(value any, _ FieldDataTypeOptions) (any, error) {
	return toString(value)
}

type fieldDataTypeUrl struct{ fieldDataTypeBase }

func (this fieldDataTypeUrl) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeUrl) Validate(value any) (any, *ft.ClientErrorItem) {
	if err := ValidateNotEmpty(value); err != nil {
		return nil, err
	}
	if err := ValidateUrl(value); err != nil {
		return nil, err
	}
	return value, nil
}

func (this fieldDataTypeUrl) TryConvert(value any, _ FieldDataTypeOptions) (any, error) {
	return toString(value)
}

type fieldDataTypeUlid struct{ fieldDataTypeBase }

func (this fieldDataTypeUlid) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeUlid) Validate(value any) (any, *ft.ClientErrorItem) {
	s, err := toString(value)
	if err != nil {
		return nil, errIncompatibleDataType()
	}
	if len(s) != model.MODEL_RULE_ULID_LENGTH {
		return nil, &ft.ClientErrorItem{
			Key:     "invalid_ulid_length",
			Message: "ulid must be {{.length}} characters",
			Vars:    map[string]any{"length": model.MODEL_RULE_ULID_LENGTH},
		}
	}
	if _, err = ulid.Parse(s); err != nil {
		return nil, &ft.ClientErrorItem{Key: "invalid_ulid", Message: "invalid ulid format", Vars: nil}
	}
	return value, nil
}

func (this fieldDataTypeUlid) TryConvert(value any, _ FieldDataTypeOptions) (any, error) {
	return toString(value)
}

type fieldDataTypeUuid struct{ fieldDataTypeBase }

func (this fieldDataTypeUuid) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeUuid) Validate(value any) (any, *ft.ClientErrorItem) {
	if err := ValidateNotEmpty(value); err != nil {
		return nil, err
	}
	if err := ValidateUuid(value); err != nil {
		return nil, err
	}
	return value, nil
}

func (this fieldDataTypeUuid) TryConvert(value any, _ FieldDataTypeOptions) (any, error) {
	return toString(value)
}

// --- Numeric types ---

type fieldDataTypeInteger struct{ fieldDataTypeBase }

func (this fieldDataTypeInteger) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeInteger) Validate(value any) (any, *ft.ClientErrorItem) {
	_, err := toInt64(value)
	if err != nil {
		return nil, errIncompatibleDataType()
	}
	return value, nil
}

func (this fieldDataTypeInteger) TryConvert(value any, _ FieldDataTypeOptions) (any, error) {
	return toInt64(value)
}

type fieldDataTypeFloat struct{ fieldDataTypeBase }

func (this fieldDataTypeFloat) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeFloat) Validate(value any) (any, *ft.ClientErrorItem) {
	_, err := toFloat64(value, this.options)
	if err != nil {
		return nil, errIncompatibleDataType()
	}
	return value, nil
}

func (this fieldDataTypeFloat) TryConvert(value any, options FieldDataTypeOptions) (any, error) {
	return toFloat64(value, options)
}

type fieldDataTypeBoolean struct{ fieldDataTypeBase }

func (this fieldDataTypeBoolean) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeBoolean) Validate(value any) (any, *ft.ClientErrorItem) {
	_, err := toBool(value)
	if err != nil {
		return nil, errIncompatibleDataType()
	}
	return value, nil
}

func (this fieldDataTypeBoolean) TryConvert(value any, _ FieldDataTypeOptions) (any, error) {
	return toBool(value)
}

// --- Date/Time types ---

type fieldDataTypeDate struct{ fieldDataTypeBase }

func (this fieldDataTypeDate) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeDate) Validate(value any) (any, *ft.ClientErrorItem) {
	_, err := toDate(value)
	if err != nil {
		return nil, errIncompatibleDataType()
	}
	return value, nil
}

func (this fieldDataTypeDate) TryConvert(value any, _ FieldDataTypeOptions) (any, error) {
	return toDate(value)
}

type fieldDataTypeTime struct{ fieldDataTypeBase }

func (this fieldDataTypeTime) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeTime) Validate(value any) (any, *ft.ClientErrorItem) {
	_, err := toTime(value)
	if err != nil {
		return nil, errIncompatibleDataType()
	}
	return value, nil
}

func (this fieldDataTypeTime) TryConvert(value any, _ FieldDataTypeOptions) (any, error) {
	return toTime(value)
}

type fieldDataTypeDateTime struct{ fieldDataTypeBase }

func (this fieldDataTypeDateTime) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeDateTime) Validate(value any) (any, *ft.ClientErrorItem) {
	_, err := toDateTime(value)
	if err != nil {
		return nil, errIncompatibleDataType()
	}
	return value, nil
}

func (this fieldDataTypeDateTime) TryConvert(value any, _ FieldDataTypeOptions) (any, error) {
	return toDateTime(value)
}

// --- Enum types ---

type fieldDataTypeEnumString struct{ fieldDataTypeBase }

func (this fieldDataTypeEnumString) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeEnumString) Validate(value any) (any, *ft.ClientErrorItem) {
	allowed := getEnumStringValues(this.options)
	if len(allowed) == 0 {
		_, err := toString(value)
		if err != nil {
			return nil, errIncompatibleDataType()
		}
		return value, nil
	}
	allowedAny := make([]any, len(allowed))
	for i, s := range allowed {
		allowedAny[i] = s
	}
	if err := ValidateOneOf(value, allowedAny); err != nil {
		return nil, err
	}
	return value, nil
}

func (this fieldDataTypeEnumString) TryConvert(value any, options FieldDataTypeOptions) (any, error) {
	return toString(value)
}

type fieldDataTypeEnumNumber struct{ fieldDataTypeBase }

func (this fieldDataTypeEnumNumber) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeEnumNumber) Validate(value any) (any, *ft.ClientErrorItem) {
	allowed := getEnumNumberValues(this.options)
	if len(allowed) == 0 {
		_, err := toInt64(value)
		if err != nil {
			return nil, errIncompatibleDataType()
		}
		return value, nil
	}
	allowedAny := make([]any, len(allowed))
	for i, n := range allowed {
		allowedAny[i] = n
	}
	if err := ValidateOneOf(value, allowedAny); err != nil {
		return nil, err
	}
	return value, nil
}

func (this fieldDataTypeEnumNumber) TryConvert(value any, options FieldDataTypeOptions) (any, error) {
	return toInt64(value)
}

// --- Nikki custom types ---

type fieldDataTypeEtag struct{ fieldDataTypeBase }

func (this fieldDataTypeEtag) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeEtag) Validate(value any) (any, *ft.ClientErrorItem) {
	if err := ValidateNotEmpty(value); err != nil {
		return nil, err
	}
	if err := ValidateLength(value, []int{model.MODEL_RULE_ETAG_MIN_LENGTH, model.MODEL_RULE_ETAG_MAX_LENGTH}); err != nil {
		return nil, err
	}
	return value, nil
}

func (this fieldDataTypeEtag) TryConvert(value any, _ FieldDataTypeOptions) (any, error) {
	return toString(value)
}

type fieldDataTypeLangJson struct{ fieldDataTypeBase }

func (this fieldDataTypeLangJson) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeLangJson) Validate(value any) (any, *ft.ClientErrorItem) {
	var lj model.LangJson
	switch x := value.(type) {
	case model.LangJson:
		if err := ValidateNotEmpty(x); err != nil {
			return nil, err
		}
		lj = x
	case *model.LangJson:
		if x == nil {
			return nil, &ft.ClientErrorItem{Key: "lang_json_nil_required", Message: "langJson cannot be nil", Vars: nil}
		}
		if err := ValidateNotEmpty(*x); err != nil {
			return nil, err
		}
		lj = *x
	case map[string]string:
		if err := ValidateNotEmpty(model.LangJson(x)); err != nil {
			return nil, err
		}
		lj = model.LangJson(x)
	default:
		return nil, &ft.ClientErrorItem{
			Key:     "incompatible_data_type",
			Message: "langJson expects map[LanguageCode]string",
			Vars:    nil,
		}
	}
	sanitized, _, err := lj.SanitizeClone(
		getLangJsonWhitelist(this.options),
		this.options[FieldDataTypeOptSanitizeType] == SanitizeTypeHtml,
	)
	if err != nil {
		return nil, &ft.ClientErrorItem{Key: "lang_json_sanitize_failed", Message: err.Error(), Vars: nil}
	}
	return *sanitized, nil
}

func (this fieldDataTypeLangJson) TryConvert(value any, _ FieldDataTypeOptions) (any, error) {
	switch v := value.(type) {
	case model.LangJson:
		return v, nil
	case *model.LangJson:
		if v == nil {
			return nil, errors.New("langJson cannot be nil")
		}
		return *v, nil
	case map[string]string:
		return model.LangJson(v), nil
	default:
		return nil, errors.Errorf("cannot convert %T to LangJson", value)
	}
}

type fieldDataTypeLangCode struct{ fieldDataTypeBase }

func (this fieldDataTypeLangCode) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeLangCode) Validate(value any) (any, *ft.ClientErrorItem) {
	s, err := toString(value)
	if err != nil {
		return nil, errIncompatibleDataType()
	}
	if s != model.LabelRefLanguageCode && !model.IsBCP47LanguageCode(s) {
		return nil, &ft.ClientErrorItem{
			Key:     "invalid_language_code",
			Message: "must be a valid BCP47-compliant language code with region part",
			Vars:    nil,
		}
	}
	return value, nil
}

func (this fieldDataTypeLangCode) TryConvert(value any, _ FieldDataTypeOptions) (any, error) {
	s, err := toString(value)
	if err != nil {
		return nil, err
	}
	canonical, err := model.ToBCP47LanguageCode(s)
	if err != nil {
		return nil, err
	}
	return canonical, nil
}

type fieldDataTypeModelId struct{ fieldDataTypeBase }

func (this fieldDataTypeModelId) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeModelId) Validate(value any) (any, *ft.ClientErrorItem) {
	if err := ValidateLength(value, []int{model.MODEL_RULE_ULID_LENGTH, model.MODEL_RULE_ULID_LENGTH}); err != nil {
		return nil, err
	}
	return value, nil
}

func (this fieldDataTypeModelId) TryConvert(value any, _ FieldDataTypeOptions) (any, error) {
	return toString(value)
}

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type fieldDataTypeSlug struct{ fieldDataTypeBase }

func (this fieldDataTypeSlug) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeSlug) Validate(value any) (any, *ft.ClientErrorItem) {
	if err := ValidateNotEmpty(value); err != nil {
		return nil, err
	}
	if err := ValidateLength(value, []int{1, model.MODEL_RULE_SHORT_NAME_LENGTH}); err != nil {
		return nil, err
	}
	if err := validatePattern(value, slugRegex); err != nil {
		return nil, err
	}
	return value, nil
}

func (this fieldDataTypeSlug) TryConvert(value any, _ FieldDataTypeOptions) (any, error) {
	s, err := toString(value)
	if err != nil {
		return nil, err
	}
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "-")
	return s, nil
}

func firstFieldDataTypeOptions(opts []FieldDataTypeOptions) FieldDataTypeOptions {
	if len(opts) == 0 {
		return nil
	}
	return opts[0]
}

// --- Helpers ---

func toString(value any) (string, error) {
	if value == nil {
		return "", errors.New("value cannot be nil")
	}
	switch v := value.(type) {
	case string:
		return v, nil
	case *string:
		if v == nil {
			return "", errors.New("value cannot be nil")
		}
		return *v, nil
	case fmt.Stringer:
		return v.String(), nil
	default:
		return fmt.Sprint(value), nil
	}
}

func toInt64(value any) (int64, error) {
	if value == nil {
		return 0, errors.New("value cannot be nil")
	}
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case uint:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		return int64(v), nil
	case float32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, errors.Errorf("cannot convert %T to integer", value)
	}
}

func toFloat64(value any, options FieldDataTypeOptions) (float64, error) {
	if value == nil {
		return 0, errors.New("value cannot be nil")
	}
	var f float64
	switch v := value.(type) {
	case float32:
		f = float64(v)
	case float64:
		f = v
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		f = reflect.ValueOf(v).Convert(reflect.TypeOf(float64(0))).Float()
	case string:
		var err error
		f, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, err
		}
	default:
		return 0, errors.Errorf("cannot convert %T to float", value)
	}
	precision := getPrecision(options)
	if precision >= 0 {
		mult := 1.0
		for i := 0; i < precision; i++ {
			mult *= 10
		}
		f = float64(int64(f*mult+0.5)) / mult
	}
	return f, nil
}

func toBool(value any) (bool, error) {
	if value == nil {
		return false, errors.New("value cannot be nil")
	}
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		s := strings.ToLower(strings.TrimSpace(v))
		if s == "true" || s == "1" || s == "yes" {
			return true, nil
		}
		if s == "false" || s == "0" || s == "no" {
			return false, nil
		}
		return false, errors.Errorf("cannot parse '%s' as boolean", v)
	default:
		return false, errors.Errorf("cannot convert %T to boolean", value)
	}
}

func toDate(value any) (time.Time, error) {
	if value == nil {
		return time.Time{}, errors.New("value cannot be nil")
	}
	switch v := value.(type) {
	case time.Time:
		return v, nil
	case *time.Time:
		if v == nil {
			return time.Time{}, errors.New("value cannot be nil")
		}
		return *v, nil
	case string:
		return time.Parse("2006-01-02", v)
	default:
		return time.Time{}, errors.Errorf("cannot convert %T to date", value)
	}
}

func toTime(value any) (time.Time, error) {
	if value == nil {
		return time.Time{}, errors.New("value cannot be nil")
	}
	switch v := value.(type) {
	case time.Time:
		return v, nil
	case *time.Time:
		if v == nil {
			return time.Time{}, errors.New("value cannot be nil")
		}
		return *v, nil
	case string:
		return time.Parse("15:04:05", v)
	default:
		return time.Time{}, errors.Errorf("cannot convert %T to time", value)
	}
}

func toDateTime(value any) (time.Time, error) {
	if value == nil {
		return time.Time{}, errors.New("value cannot be nil")
	}
	switch v := value.(type) {
	case time.Time:
		return v, nil
	case *time.Time:
		if v == nil {
			return time.Time{}, errors.New("value cannot be nil")
		}
		return *v, nil
	case string:
		formats := []string{
			time.RFC3339,
			"2006-01-02T15:04:05Z07:00",
			"2006-01-02 15:04:05",
			"2006-01-02",
		}
		for _, layout := range formats {
			if t, err := time.Parse(layout, v); err == nil {
				return t, nil
			}
		}
		return time.Time{}, errors.Errorf("cannot parse '%s' as datetime", v)
	default:
		return time.Time{}, errors.Errorf("cannot convert %T to datetime", value)
	}
}

func getEnumStringValues(options FieldDataTypeOptions) []string {
	if options == nil {
		return nil
	}
	raw, ok := options[FieldDataTypeOptEnumValues]
	if !ok || raw == nil {
		return nil
	}
	switch v := raw.(type) {
	case []string:
		return v
	case []any:
		result := make([]string, 0, len(v))
		for _, item := range v {
			result = append(result, fmt.Sprint(item))
		}
		return result
	default:
		return nil
	}
}

func getEnumNumberValues(options FieldDataTypeOptions) []int64 {
	if options == nil {
		return nil
	}
	raw, ok := options[FieldDataTypeOptEnumValues]
	if !ok || raw == nil {
		return nil
	}
	switch v := raw.(type) {
	case []int:
		result := make([]int64, len(v))
		for i, n := range v {
			result[i] = int64(n)
		}
		return result
	case []int64:
		return v
	case []any:
		result := make([]int64, 0, len(v))
		for _, item := range v {
			n, err := toInt64(item)
			if err != nil {
				return nil
			}
			result = append(result, n)
		}
		return result
	default:
		return nil
	}
}

func getLangJsonWhitelist(options FieldDataTypeOptions) []model.LanguageCode {
	if options == nil {
		return nil
	}
	raw, ok := options[FieldDataTypeOptLangJsonWhitelist]
	if !ok || raw == nil {
		return nil
	}
	switch v := raw.(type) {
	case []string:
		result := make([]model.LanguageCode, len(v))
		for i, s := range v {
			result[i] = model.LanguageCode(s)
		}
		return result
	case []any:
		result := make([]model.LanguageCode, 0, len(v))
		for _, item := range v {
			result = append(result, model.LanguageCode(fmt.Sprint(item)))
		}
		return result
	default:
		return nil
	}
}

func getPrecision(options FieldDataTypeOptions) int {
	if options == nil {
		return -1
	}
	raw, ok := options[FieldDataTypeOptPrecision]
	if !ok || raw == nil {
		return -1
	}
	switch v := raw.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		n, _ := strconv.Atoi(v)
		return n
	default:
		return -1
	}
}

func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func containsInt64(slice []int64, n int64) bool {
	for _, v := range slice {
		if v == n {
			return true
		}
	}
	return false
}

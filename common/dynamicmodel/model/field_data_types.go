package model

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"
	"go.bryk.io/pkg/ulid"

	"github.com/sky-as-code/nikki-erp/common/defense"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
)

// FieldDataType defines the interface for dynamic field data types.
// Validate returns (validatedValue, nil) on success or (nil, ValidationError) on failure.
// Validate always runs TryConvert first to coerce the payload to the concrete storage type,
// then applies type-specific rules. Options are embedded in the data type; both use them internally.
type FieldDataType interface {
	ArrayType() FieldDataType
	DefaultValue() value
	IsArray() bool
	Options() FieldDataTypeOptions
	String() string
	TryConvert(val any, options FieldDataTypeOptions) (value, error)
	Validate(val value) (value, *ft.ClientErrorItem)
}

// --- Factory functions (replace package-level vars) ---
// Scalar types are created by default. Use .ArrayType() to get array variant.

func FieldDataTypeEmail() FieldDataType {
	return fieldDataTypeEmail{fieldDataTypeBase{name: "email", options: nil}}
}

func FieldDataTypePhone() FieldDataType {
	return fieldDataTypePhone{fieldDataTypeBase{name: "phone", options: nil}}
}

func FieldDataTypeString(minLength int, maxLength int, sanitizeType ...SanitizeType) FieldDataType {
	validateRangeOrPanic(minLength, maxLength, "FieldDataTypeString")
	st := SanitizeTypePlainText
	if len(sanitizeType) > 0 && sanitizeType[0] != "" {
		st = sanitizeType[0]
	}
	opts := FieldDataTypeOptions{
		FieldDataTypeOptSanitizeType: st,
		FieldDataTypeOptLength:       []int{minLength, maxLength},
	}
	return fieldDataTypeString{fieldDataTypeBase{name: "string", options: opts}}
}

func FieldDataTypeSecret() FieldDataType {
	return fieldDataTypeSecret{fieldDataTypeBase{
		name: "secret",
		options: FieldDataTypeOptions{
			FieldDataTypeOptSanitizeType: SanitizeTypeNone,
		},
	}}
}

func FieldDataTypeUrl() FieldDataType {
	return fieldDataTypeUrl{fieldDataTypeBase{
		name: "url",
		options: FieldDataTypeOptions{
			FieldDataTypeOptLength: []int{model.MODEL_RULE_URL_LENGTH_MIN, model.MODEL_RULE_URL_LENGTH_MAX},
		},
	}}
}

func FieldDataTypeUlid() FieldDataType {
	return fieldDataTypeUlid{fieldDataTypeBase{name: "ulid", options: nil}}
}

func FieldDataTypeUuid() FieldDataType {
	return fieldDataTypeUuid{fieldDataTypeBase{name: "uuid", options: nil}}
}

func FieldDataTypeInt64(min int64, max int64) FieldDataType {
	validateRangeOrPanic(min, max, "FieldDataTypeInt64")
	opts := FieldDataTypeOptions{FieldDataTypeOptRange: []int64{min, max}}
	return fieldDataTypeInt64{fieldDataTypeBase{name: "int64", options: opts}}
}

func FieldDataTypeInt(min int, max int) FieldDataType {
	validateRangeOrPanic(min, max, "FieldDataTypeInt")
	opts := FieldDataTypeOptions{FieldDataTypeOptRange: []int{min, max}}
	return fieldDataTypeInt{fieldDataTypeBase{name: "int", options: opts}}
}

func FieldDataTypeDecimal(min string, max string, scale uint) FieldDataType {
	validateDecimalRangeAndScaleOrPanic(min, max, scale)
	opts := FieldDataTypeOptions{
		FieldDataTypeOptRange: []string{min, max},
		FieldDataTypeOptScale: scale,
	}
	return fieldDataTypeDecimal{fieldDataTypeBase{name: "decimal", options: opts}}
}

func FieldDataTypeBoolean() FieldDataType {
	return fieldDataTypeBoolean{fieldDataTypeBase{name: "boolean", options: nil}}
}

func FieldDataTypeDate() FieldDataType {
	return fieldDataTypeDate{fieldDataTypeBase{name: "nikkiDate", options: nil}}
}

func FieldDataTypeTime() FieldDataType {
	return fieldDataTypeTime{fieldDataTypeBase{name: "nikkiTime", options: nil}}
}

func FieldDataTypeDateTime() FieldDataType {
	return fieldDataTypeDateTime{fieldDataTypeBase{name: "nikkiDateTime", options: nil}}
}

func FieldDataTypeEnumString(enumValues []string) FieldDataType {
	opts := FieldDataTypeOptions{FieldDataTypeOptEnumValues: enumValues}
	return fieldDataTypeEnumString{fieldDataTypeBase{name: "enumString", options: opts}}
}

func FieldDataTypeEnumInteger(enumValues []int64) FieldDataType {
	opts := FieldDataTypeOptions{FieldDataTypeOptEnumValues: enumValues}
	return fieldDataTypeEnumInteger{fieldDataTypeBase{name: "enumInteger", options: opts}}
}

func FieldDataTypeEtag() FieldDataType {
	return fieldDataTypeEtag{fieldDataTypeBase{
		name: "nikkiEtag",
		options: FieldDataTypeOptions{
			FieldDataTypeOptLength: []int{model.MODEL_RULE_ETAG_MIN_LENGTH, model.MODEL_RULE_ETAG_MAX_LENGTH},
		},
	}}
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

func FieldDataTypeSlug() FieldDataType {
	return fieldDataTypeSlug{fieldDataTypeBase{
		name: "nikkiSlug",
		options: FieldDataTypeOptions{
			FieldDataTypeOptLength: []int{model.MODEL_RULE_SLUG_LENGTH_MIN, model.MODEL_RULE_SLUG_LENGTH_MAX},
		},
	}}
}

// FieldDataTypeModel represents a virtual/implicit field that holds a related model or slice of models.
// It is not persisted as a DB column; it is used for graph traversal and API response expansion.
func FieldDataTypeModel() FieldDataType {
	return fieldDataTypeModel{fieldDataTypeBase{name: "model", options: nil}}
}

func IsFieldDataTypeModel(dt FieldDataType) bool {
	return dt.String() == "model"
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

func (this fieldDataTypeBase) DefaultValue() value {
	return Value(nil)
}

type fieldDataTypeModel struct{ fieldDataTypeBase }

func (this fieldDataTypeModel) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeModel) DefaultValue() value {
	return Value(nil)
}

func (this fieldDataTypeModel) validateScalar(c value) (value, *ft.ClientErrorItem) {
	return c, nil
}

func (this fieldDataTypeModel) Validate(value value) (value, *ft.ClientErrorItem) {
	if this.isArray {
		return validateArrayAfterTryConvert(this, value, this.validateScalar)
	}
	return validateScalarAfterTryConvert(this, value, this.validateScalar)
}

func (this fieldDataTypeModel) TryConvert(val any, _ FieldDataTypeOptions) (value, error) {
	if val == nil {
		return Value(nil), nil
	}
	rv := reflect.ValueOf(val)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return Value(nil), nil
		}
		val = rv.Elem().Interface()
	}
	return Value(val), nil
}

// --- String-like types ---

type fieldDataTypeEmail struct{ fieldDataTypeBase }

func (this fieldDataTypeEmail) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeEmail) Validate(val value) (value, *ft.ClientErrorItem) {
	if this.isArray {
		return validateArrayAfterTryConvert(this, val, this.validateScalar)
	}
	return validateScalarAfterTryConvert(this, val, this.validateScalar)
}

func (this fieldDataTypeEmail) validateScalar(val value) (value, *ft.ClientErrorItem) {
	sanitized, clientErr := validateStringBase(val, this.options)
	if clientErr != nil {
		return Value(nil), clientErr
	}
	if ve := ValidateEmail((*sanitized.Get()).(string)); ve != nil {
		return Value(nil), ve
	}
	return sanitized, nil
}

func (this fieldDataTypeEmail) TryConvert(val any, _ FieldDataTypeOptions) (value, error) {
	_, sameType := val.(string)
	_, ptrSameType := val.(*string)
	if sameType || ptrSameType {
		return Value(val), nil
	}
	str, err := toString(val)
	if err != nil {
		return Value(nil), err
	}
	return Value(str), nil
}

type fieldDataTypePhone struct{ fieldDataTypeBase }

func (this fieldDataTypePhone) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypePhone) Validate(val value) (value, *ft.ClientErrorItem) {
	if this.isArray {
		return validateArrayAfterTryConvert(this, val, this.validateScalar)
	}
	return validateScalarAfterTryConvert(this, val, this.validateScalar)
}

func (this fieldDataTypePhone) validateScalar(val value) (value, *ft.ClientErrorItem) {
	return validateStringBase(val, this.options)
}

func (this fieldDataTypePhone) TryConvert(val any, _ FieldDataTypeOptions) (value, error) {
	_, sameType := val.(string)
	_, ptrSameType := val.(*string)
	if sameType || ptrSameType {
		return Value(val), nil
	}
	str, err := toString(val)
	if err != nil {
		return Value(nil), err
	}
	return Value(str), nil
}

type fieldDataTypeString struct{ fieldDataTypeBase }

func (this fieldDataTypeString) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeString) Validate(val value) (value, *ft.ClientErrorItem) {
	if this.isArray {
		return validateArrayAfterTryConvert(this, val, this.validateScalar)
	}
	return validateScalarAfterTryConvert(this, val, this.validateScalar)
}

func (this fieldDataTypeString) validateScalar(val value) (value, *ft.ClientErrorItem) {
	return validateStringBase(val, this.options)
}

func validateStringBase(val value, options FieldDataTypeOptions) (value, *ft.ClientErrorItem) {
	if val.Get() == nil {
		return Value(nil), NewInvalidDataTypeErr("")
	}
	raw := *val.Get()
	s, err := toString(raw)
	if err != nil {
		return Value(nil), NewInvalidDataTypeErr("")
	}
	sanitized, clientErr := sanitizeStringValue(raw, options)
	if clientErr != nil {
		return Value(nil), clientErr
	}
	if clientErr := validateStringLength(s, options); clientErr != nil {
		return Value(nil), clientErr
	}
	var out string
	switch v := sanitized.(type) {
	case string:
		out = v
	case *string:
		if v == nil {
			return Value(nil), NewInvalidDataTypeErr("")
		}
		out = *v
	default:
		return Value(nil), NewInvalidDataTypeErr("")
	}
	return Value(out), nil
}

func validateStringLength(s string, options FieldDataTypeOptions) *ft.ClientErrorItem {
	if options == nil {
		return nil
	}
	opts, hasLimits := options[FieldDataTypeOptLength]
	if !hasLimits {
		return nil
	}
	limits := opts.([]int)
	length := len(s)
	min := limits[0]
	max := limits[1]
	if length < min || length > max {
		return ft.NewAnonymousValidationError(
			ft.ErrorKey("err_invalid_string_length"),
			"string length must be between {{.min}} and {{.max}}",
			map[string]any{"min": min, "max": max},
		)
	}
	return nil
}

func sanitizeStringValue(value any, options FieldDataTypeOptions) (any, *ft.ClientErrorItem) {
	if options == nil {
		return value, nil
	}
	raw, ok := options[FieldDataTypeOptSanitizeType]
	if !ok || raw == nil {
		return value, nil
	}
	st, ok := raw.(SanitizeType)
	if !ok {
		return value, nil
	}
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
	case SanitizeTypeNone:
		return s
	case SanitizeTypeHtml:
		return defense.SanitizeRichText(s)
	case SanitizeTypePlainText:
		return defense.SanitizePlainText(s, true)
	default:
		return s
	}
}

func (this fieldDataTypeString) TryConvert(val any, _ FieldDataTypeOptions) (value, error) {
	_, sameType := val.(string)
	_, ptrSameType := val.(*string)
	if sameType || ptrSameType {
		return Value(val), nil
	}
	str, err := toString(val)
	if err != nil {
		return Value(nil), err
	}
	return Value(str), nil
}

type fieldDataTypeSecret struct{ fieldDataTypeBase }

func (this fieldDataTypeSecret) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeSecret) Validate(val value) (value, *ft.ClientErrorItem) {
	if this.isArray {
		return validateArrayAfterTryConvert(this, val, this.validateScalar)
	}
	return validateScalarAfterTryConvert(this, val, this.validateScalar)
}

func (this fieldDataTypeSecret) validateScalar(val value) (value, *ft.ClientErrorItem) {
	return validateStringBase(val, this.options)
}

func (this fieldDataTypeSecret) TryConvert(val any, _ FieldDataTypeOptions) (value, error) {
	_, sameType := val.(string)
	_, ptrSameType := val.(*string)
	if sameType || ptrSameType {
		return Value(val), nil
	}
	str, err := toString(val)
	if err != nil {
		return Value(nil), err
	}
	return Value(str), nil
}

type fieldDataTypeUrl struct{ fieldDataTypeBase }

func (this fieldDataTypeUrl) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeUrl) Validate(val value) (value, *ft.ClientErrorItem) {
	if this.isArray {
		return validateArrayAfterTryConvert(this, val, this.validateScalar)
	}
	return validateScalarAfterTryConvert(this, val, this.validateScalar)
}

func (this fieldDataTypeUrl) validateScalar(val value) (value, *ft.ClientErrorItem) {
	sanitized, clientErr := validateStringBase(val, this.options)
	if clientErr != nil {
		return Value(nil), clientErr
	}
	return sanitized, ValidateUrl((*sanitized.Get()).(string))
}

func (this fieldDataTypeUrl) TryConvert(val any, _ FieldDataTypeOptions) (value, error) {
	_, sameType := val.(string)
	_, ptrSameType := val.(*string)
	if sameType || ptrSameType {
		return Value(val), nil
	}
	str, err := toString(val)
	if err != nil {
		return Value(nil), err
	}
	return Value(str), nil
}

type fieldDataTypeUlid struct{ fieldDataTypeBase }

func (this fieldDataTypeUlid) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeUlid) DefaultValue() value {
	id, err := model.NewId()
	if err != nil {
		panic(err)
	}
	return Value(*id)
}

func (this fieldDataTypeUlid) Validate(val value) (value, *ft.ClientErrorItem) {
	if this.isArray {
		return validateArrayAfterTryConvert(this, val, this.validateScalar)
	}
	return validateScalarAfterTryConvert(this, val, this.validateScalar)
}

func (this fieldDataTypeUlid) validateScalar(val value) (value, *ft.ClientErrorItem) {
	sanitized, clientErr := validateStringBase(val, this.options)
	if clientErr != nil {
		return Value(nil), clientErr
	}
	s := (*sanitized.Get()).(string)
	if len(s) != model.MODEL_RULE_ULID_LENGTH {
		return Value(nil), NewInvalidDataTypeErr("")
	}
	if _, err := ulid.Parse(s); err != nil {
		return Value(nil), NewInvalidDataTypeErr("")
	}
	return sanitized, nil
}

func (this fieldDataTypeUlid) TryConvert(val any, _ FieldDataTypeOptions) (value, error) {
	_, sameType := val.(string)
	_, ptrSameType := val.(*string)
	if sameType || ptrSameType {
		return Value(val), nil
	}
	str, err := toString(val)
	if err != nil {
		return Value(nil), err
	}
	return Value(str), nil
}

type fieldDataTypeUuid struct{ fieldDataTypeBase }

func (this fieldDataTypeUuid) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeUuid) Validate(val value) (value, *ft.ClientErrorItem) {
	if this.isArray {
		return validateArrayAfterTryConvert(this, val, this.validateScalar)
	}
	return validateScalarAfterTryConvert(this, val, this.validateScalar)
}

func (this fieldDataTypeUuid) validateScalar(val value) (value, *ft.ClientErrorItem) {
	sanitized, clientErr := validateStringBase(val, this.options)
	if clientErr != nil {
		return Value(nil), clientErr
	}
	if !ValidateUuid((*sanitized.Get()).(string)) {
		return Value(nil), NewInvalidDataTypeErr("")
	}
	return sanitized, nil
}

func (this fieldDataTypeUuid) TryConvert(val any, _ FieldDataTypeOptions) (value, error) {
	_, sameType := val.(string)
	_, ptrSameType := val.(*string)
	if sameType || ptrSameType {
		return Value(val), nil
	}
	str, err := toString(val)
	if err != nil {
		return Value(nil), err
	}
	return Value(str), nil
}

// --- Numeric types ---

type fieldDataTypeInt64 struct{ fieldDataTypeBase }

func (this fieldDataTypeInt64) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeInt64) DefaultValue() value {
	return Value(int64(0))
}

func (this fieldDataTypeInt64) Validate(val value) (value, *ft.ClientErrorItem) {
	if this.isArray {
		return validateArrayAfterTryConvert(this, val, this.validateScalar)
	}
	return validateScalarAfterTryConvert(this, val, this.validateScalar)
}

func (this fieldDataTypeInt64) validateScalar(val value) (value, *ft.ClientErrorItem) {
	if val.Get() == nil {
		return Value(nil), NewInvalidDataTypeErr("")
	}
	n, ok := (*val.Get()).(int64)
	if !ok {
		return Value(nil), NewInvalidDataTypeErr("")
	}
	limits := getInt64Range(this.options)
	if len(limits) == 2 && (n < limits[0] || n > limits[1]) {
		return Value(nil), ft.NewAnonymousValidationError(
			ft.ErrorKey("err_invalid_number_range"),
			"value must be between {{.min}} and {{.max}}",
			map[string]any{"min": limits[0], "max": limits[1]},
		)
	}
	return val, nil
}

func (this fieldDataTypeInt64) TryConvert(val any, _ FieldDataTypeOptions) (value, error) {
	_, sameType := val.(int64)
	_, ptrSameType := val.(*int64)
	if sameType || ptrSameType {
		return Value(val), nil
	}
	result, err := toInt64(val)
	if err != nil {
		return Value(nil), err
	}
	return Value(result), nil
}

type fieldDataTypeInt struct{ fieldDataTypeBase }

func (this fieldDataTypeInt) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeInt) DefaultValue() value {
	return Value(int(0))
}

func (this fieldDataTypeInt) Validate(val value) (value, *ft.ClientErrorItem) {
	if this.isArray {
		return validateArrayAfterTryConvert(this, val, this.validateScalar)
	}
	return validateScalarAfterTryConvert(this, val, this.validateScalar)
}

func (this fieldDataTypeInt) validateScalar(val value) (value, *ft.ClientErrorItem) {
	if val.Get() == nil {
		return Value(nil), NewInvalidDataTypeErr("")
	}
	n, ok := (*val.Get()).(int)
	if !ok {
		return Value(nil), NewInvalidDataTypeErr("")
	}
	limits := getIntRange(this.options)
	if len(limits) == 2 && (n < limits[0] || n > limits[1]) {
		return Value(nil), ft.NewAnonymousValidationError(
			ft.ErrorKey("err_invalid_number_range"),
			"value must be between {{.min}} and {{.max}}",
			map[string]any{"min": limits[0], "max": limits[1]},
		)
	}
	return val, nil
}

func (this fieldDataTypeInt) TryConvert(val any, _ FieldDataTypeOptions) (value, error) {
	_, sameType := val.(int)
	_, ptrSameType := val.(*int)
	if sameType || ptrSameType {
		return Value(val), nil
	}
	result, err := toInt(val)
	if err != nil {
		return Value(nil), err
	}
	return Value(result), nil
}

type fieldDataTypeDecimal struct{ fieldDataTypeBase }

func (this fieldDataTypeDecimal) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeDecimal) DefaultValue() value {
	return Value(nil)
}

func (this fieldDataTypeDecimal) Validate(val value) (value, *ft.ClientErrorItem) {
	if this.isArray {
		return validateArrayAfterTryConvert(this, val, this.validateScalar)
	}
	return validateScalarAfterTryConvert(this, val, this.validateScalar)
}

func (this fieldDataTypeDecimal) validateScalar(val value) (value, *ft.ClientErrorItem) {
	if val.Get() == nil {
		return Value(nil), NewInvalidDataTypeErr("")
	}
	n, err := toDecimal(*val.Get())
	if err != nil {
		return Value(nil), NewInvalidDataTypeErr("")
	}
	minMax, err := getDecimalRange(this.options)
	if err != nil {
		return Value(nil), NewInvalidDataTypeErr("")
	}
	if len(minMax) == 2 && (n.LessThan(minMax[0]) || n.GreaterThan(minMax[1])) {
		return Value(nil), ft.NewAnonymousValidationError(
			ft.ErrorKey("err_invalid_number_range"),
			"value must be between {{.min}} and {{.max}}",
			map[string]any{"min": minMax[0].String(), "max": minMax[1].String()},
		)
	}
	scaled := applyDecimalScale(n, this.options)
	return Value(scaled), nil
}

func (this fieldDataTypeDecimal) TryConvert(val any, _ FieldDataTypeOptions) (value, error) {
	switch typed := val.(type) {
	case decimal.Decimal:
		return Value(typed), nil
	case *decimal.Decimal:
		if typed == nil {
			return Value(nil), errors.New("fieldDataTypeDecimal.TryConvert: value cannot be nil")
		}
		return Value(typed), nil
	case string:
		result, err := decimal.NewFromString(typed)
		if err != nil {
			return Value(nil), err
		}
		scaled := applyDecimalScale(result, this.options)
		return Value(scaled), nil
	default:
		return Value(nil), errors.New("fieldDataTypeDecimal.TryConvert: value must be decimal or string")
	}
}

type fieldDataTypeBoolean struct{ fieldDataTypeBase }

func (this fieldDataTypeBoolean) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeBoolean) DefaultValue() value {
	return Value(false)
}

func (this fieldDataTypeBoolean) Validate(value value) (value, *ft.ClientErrorItem) {
	if this.isArray {
		return validateArrayAfterTryConvert(this, value, this.validateScalar)
	}
	return validateScalarAfterTryConvert(this, value, this.validateScalar)
}

func (this fieldDataTypeBoolean) validateScalar(val value) (value, *ft.ClientErrorItem) {
	return val, nil
}

func (this fieldDataTypeBoolean) TryConvert(val any, _ FieldDataTypeOptions) (value, error) {
	_, sameType := val.(bool)
	_, ptrSameType := val.(*bool)
	if sameType || ptrSameType {
		return Value(val), nil
	}
	result, err := toBool(val)
	if err != nil {
		return Value(nil), err
	}
	return Value(result), nil
}

// --- Date/Time types ---

type fieldDataTypeDate struct{ fieldDataTypeBase }

func (this fieldDataTypeDate) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeDate) DefaultValue() value {
	return Value(model.NewModelDate())
}

func (this fieldDataTypeDate) Validate(val value) (value, *ft.ClientErrorItem) {
	if this.isArray {
		return validateArrayAfterTryConvert(this, val, this.validateScalar)
	}
	return validateScalarAfterTryConvert(this, val, this.validateScalar)
}

func (this fieldDataTypeDate) validateScalar(val value) (value, *ft.ClientErrorItem) {
	return val, nil
}

func (this fieldDataTypeDate) TryConvert(val any, _ FieldDataTypeOptions) (value, error) {
	_, sameType := val.(model.ModelDate)
	_, ptrSameType := val.(*model.ModelDate)
	if sameType || ptrSameType {
		return Value(val), nil
	}
	result, err := toDate(val)
	if err != nil {
		return Value(nil), err
	}
	return Value(result), nil
}

type fieldDataTypeTime struct{ fieldDataTypeBase }

func (this fieldDataTypeTime) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeTime) DefaultValue() value {
	return Value(model.NewModelTime())
}

func (this fieldDataTypeTime) Validate(val value) (value, *ft.ClientErrorItem) {
	if this.isArray {
		return validateArrayAfterTryConvert(this, val, this.validateScalar)
	}
	return validateScalarAfterTryConvert(this, val, this.validateScalar)
}

func (this fieldDataTypeTime) validateScalar(val value) (value, *ft.ClientErrorItem) {
	return val, nil
}

func (this fieldDataTypeTime) TryConvert(val any, _ FieldDataTypeOptions) (value, error) {
	_, sameType := val.(model.ModelTime)
	_, ptrSameType := val.(*model.ModelTime)
	if sameType || ptrSameType {
		return Value(val), nil
	}
	result, err := toTime(val)
	if err != nil {
		return Value(nil), err
	}
	return Value(result), nil
}

type fieldDataTypeDateTime struct{ fieldDataTypeBase }

func (this fieldDataTypeDateTime) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeDateTime) DefaultValue() value {
	return Value(model.NewModelDateTime())
}

func (this fieldDataTypeDateTime) Validate(val value) (value, *ft.ClientErrorItem) {
	if this.isArray {
		return validateArrayAfterTryConvert(this, val, this.validateScalar)
	}
	return validateScalarAfterTryConvert(this, val, this.validateScalar)
}

func (this fieldDataTypeDateTime) validateScalar(val value) (value, *ft.ClientErrorItem) {
	return val, nil
}

func (this fieldDataTypeDateTime) TryConvert(val any, _ FieldDataTypeOptions) (value, error) {
	_, sameType := val.(model.ModelDateTime)
	_, ptrSameType := val.(*model.ModelDateTime)
	if sameType || ptrSameType {
		return Value(val), nil
	}
	result, err := toDateTime(val)
	if err != nil {
		return Value(nil), err
	}
	return Value(result), nil
}

// --- Enum types ---

type fieldDataTypeEnumString struct{ fieldDataTypeBase }

func (this fieldDataTypeEnumString) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeEnumString) Validate(val value) (value, *ft.ClientErrorItem) {
	if this.isArray {
		return validateArrayAfterTryConvert(this, val, this.validateScalar)
	}
	return validateScalarAfterTryConvert(this, val, this.validateScalar)
}

func (this fieldDataTypeEnumString) validateScalar(val value) (value, *ft.ClientErrorItem) {
	sanitized, clientErr := validateStringBase(val, this.options)
	if clientErr != nil {
		return Value(nil), clientErr
	}
	allowed := getEnumStringValues(this.options)
	if len(allowed) == 0 {
		return sanitized, nil
	}
	allowedAny := make([]any, len(allowed))
	for i, s := range allowed {
		allowedAny[i] = s
	}
	if err := ValidateOneOf((*sanitized.Get()).(string), allowedAny); err != nil {
		return Value(nil), err
	}
	return sanitized, nil
}

func (this fieldDataTypeEnumString) TryConvert(val any, _ FieldDataTypeOptions) (value, error) {
	_, sameType := val.(string)
	_, ptrSameType := val.(*string)
	if sameType || ptrSameType {
		return Value(val), nil
	}
	str, err := toString(val)
	if err != nil {
		return Value(nil), err
	}
	return Value(str), nil
}

type fieldDataTypeEnumInteger struct{ fieldDataTypeBase }

func (this fieldDataTypeEnumInteger) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeEnumInteger) DefaultValue() value {
	return Value(int64(0))
}

func (this fieldDataTypeEnumInteger) Validate(value value) (value, *ft.ClientErrorItem) {
	if this.isArray {
		return validateArrayAfterTryConvert(this, value, this.validateScalar)
	}
	return validateScalarAfterTryConvert(this, value, this.validateScalar)
}

func (this fieldDataTypeEnumInteger) validateScalar(value value) (value, *ft.ClientErrorItem) {
	allowed := getEnumNumberValues(this.options)
	if len(allowed) == 0 {
		return value, nil
	}
	allowedAny := make([]any, len(allowed))
	for i, n := range allowed {
		allowedAny[i] = n
	}
	if value.Get() == nil {
		return Value(nil), NewInvalidDataTypeErr("")
	}
	if err := ValidateOneOf(*value.Get(), allowedAny); err != nil {
		return Value(nil), err
	}
	return value, nil
}

func (this fieldDataTypeEnumInteger) TryConvert(val any, _ FieldDataTypeOptions) (value, error) {
	_, sameType := val.(int64)
	_, ptrSameType := val.(*int64)
	if sameType || ptrSameType {
		return Value(val), nil
	}
	result, err := toInt64(val)
	if err != nil {
		return Value(nil), err
	}
	return Value(result), nil
}

// --- Nikki custom types ---

type fieldDataTypeEtag struct{ fieldDataTypeBase }

func (this fieldDataTypeEtag) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeEtag) DefaultValue() value {
	return Value(*model.NewEtag())
}

func (this fieldDataTypeEtag) Validate(value value) (value, *ft.ClientErrorItem) {
	if this.isArray {
		return validateArrayAfterTryConvert(this, value, this.validateScalar)
	}
	return validateScalarAfterTryConvert(this, value, this.validateScalar)
}

func (this fieldDataTypeEtag) validateScalar(val value) (value, *ft.ClientErrorItem) {
	sanitized, clientErr := validateStringBase(val, this.options)
	if clientErr != nil {
		return Value(nil), clientErr
	}
	return sanitized, nil
}

func (this fieldDataTypeEtag) TryConvert(val any, _ FieldDataTypeOptions) (value, error) {
	_, sameType := val.(string)
	_, ptrSameType := val.(*string)
	if sameType || ptrSameType {
		return Value(val), nil
	}
	str, err := toString(val)
	if err != nil {
		return Value(nil), err
	}
	return Value(str), nil
}

type fieldDataTypeLangJson struct{ fieldDataTypeBase }

func (this fieldDataTypeLangJson) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeLangJson) DefaultValue() value {
	return Value(model.LangJson{})
}

func (this fieldDataTypeLangJson) Validate(val value) (value, *ft.ClientErrorItem) {
	if this.isArray {
		return validateArrayAfterTryConvert(this, val, this.validateScalar)
	}
	return validateScalarAfterTryConvert(this, val, this.validateScalar)
}

func (this fieldDataTypeLangJson) validateScalar(value value) (value, *ft.ClientErrorItem) {
	if value.Get() == nil {
		return Value(nil), NewInvalidDataTypeErr("")
	}
	lj, clientErr := toLangJson(*value.Get())
	if clientErr != nil {
		return Value(nil), clientErr
	}
	sanitized, _, err := lj.SanitizeClone(
		getLangJsonWhitelist(this.options),
		this.options[FieldDataTypeOptSanitizeType] == SanitizeTypePlainText,
	)
	if err != nil {
		return Value(nil), &ft.ClientErrorItem{
			Key: "lang_json_sanitize_failed", Message: err.Error(), Vars: nil,
		}
	}
	return Value(*sanitized), nil
}

func toLangJson(value any) (model.LangJson, *ft.ClientErrorItem) {
	switch x := value.(type) {
	case model.LangJson:
		if err := ValidateNotEmpty(x); err != nil {
			return model.LangJson{}, err
		}
		return x, nil
	case *model.LangJson:
		if x == nil {
			return model.LangJson{}, &ft.ClientErrorItem{
				Key: "lang_json_nil_required", Message: "langJson cannot be nil", Vars: nil,
			}
		}
		if err := ValidateNotEmpty(*x); err != nil {
			return model.LangJson{}, err
		}
		return *x, nil
	case map[string]string:
		if err := ValidateNotEmpty(model.LangJson(x)); err != nil {
			return model.LangJson{}, err
		}
		return model.LangJson(x), nil
	default:
		return model.LangJson{}, &ft.ClientErrorItem{
			Key:     "incompatible_data_type",
			Message: "langJson expects map[LanguageCode]string",
			Vars:    nil,
		}
	}
}

func (this fieldDataTypeLangJson) TryConvert(val any, _ FieldDataTypeOptions) (value, error) {
	switch v := val.(type) {
	case model.LangJson:
		return Value(v), nil
	case *model.LangJson:
		if v == nil {
			return Value(nil), errors.New("fieldDataTypeLangJson.TryConvert: langJson cannot be nil")
		}
		return Value(*v), nil
	case map[string]string:
		return Value(model.LangJson(v)), nil
	default:
		return Value(nil), errors.Errorf(
			"fieldDataTypeLangJson.TryConvert: cannot convert %T to LangJson", val,
		)
	}
}

type fieldDataTypeLangCode struct{ fieldDataTypeBase }

func (this fieldDataTypeLangCode) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeLangCode) Validate(value value) (value, *ft.ClientErrorItem) {
	if this.isArray {
		return validateArrayAfterTryConvert(this, value, this.validateScalar)
	}
	return validateScalarAfterTryConvert(this, value, this.validateScalar)
}

func (this fieldDataTypeLangCode) validateScalar(value value) (value, *ft.ClientErrorItem) {
	sanitized, clientErr := validateStringBase(value, this.options)
	if clientErr != nil {
		return Value(nil), clientErr
	}
	s := (*sanitized.Get()).(string)
	if s != model.LabelRefLanguageCode && !model.IsBCP47LanguageCode(s) {
		return Value(nil), &ft.ClientErrorItem{
			Key:     "invalid_language_code",
			Message: "must be a valid BCP47-compliant language code with region part",
			Vars:    nil,
		}
	}
	return sanitized, nil
}

func (this fieldDataTypeLangCode) TryConvert(val any, _ FieldDataTypeOptions) (value, error) {
	s, err := toString(val)
	if err != nil {
		return Value(nil), err
	}
	canonical, err := model.ToBCP47LanguageCode(s)
	if err != nil {
		return Value(nil), err
	}
	return Value(canonical), nil
}

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type fieldDataTypeSlug struct{ fieldDataTypeBase }

func (this fieldDataTypeSlug) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeSlug) Validate(value value) (value, *ft.ClientErrorItem) {
	if this.isArray {
		return validateArrayAfterTryConvert(this, value, this.validateScalar)
	}
	return validateScalarAfterTryConvert(this, value, this.validateScalar)
}

func (this fieldDataTypeSlug) validateScalar(val value) (value, *ft.ClientErrorItem) {
	sanitized, clientErr := validateStringBase(val, this.options)
	if clientErr != nil {
		return Value(nil), clientErr
	}
	if !ValidatePattern((*sanitized.Get()).(string), slugRegex) {
		return Value(nil), NewInvalidDataTypeErr("")
	}
	return sanitized, nil
}

func (this fieldDataTypeSlug) TryConvert(val any, _ FieldDataTypeOptions) (value, error) {
	s, err := toString(val)
	if err != nil {
		return Value(nil), err
	}
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "-")
	return Value(s), nil
}

// --- Helpers ---

var (
	reflectTypeInt64   = reflect.TypeOf(int64(0))
	reflectTypeInt     = reflect.TypeOf(int(0))
	reflectTypeFloat64 = reflect.TypeOf(float64(0))
	reflectTypeBool    = reflect.TypeOf(false)
)

func float64IfExactOrPtr(val any) (float64, bool) {
	if val == nil {
		return 0, false
	}
	switch x := val.(type) {
	case float64:
		return x, true
	case *float64:
		if x == nil {
			return 0, false
		}
		return *x, true
	default:
		return 0, false
	}
}

func tryConvertOrIncompatible(dt FieldDataType, raw any) (value, *ft.ClientErrorItem) {
	converted, err := dt.TryConvert(raw, dt.Options())
	if err != nil {
		return Value(nil), NewInvalidDataTypeErr("")
	}
	return converted, nil
}

func validateScalarAfterTryConvert(
	dt FieldDataType,
	val value,
	validateConverted func(value) (value, *ft.ClientErrorItem),
) (value, *ft.ClientErrorItem) {
	if val.Get() == nil {
		return Value(nil), NewInvalidDataTypeErr("")
	}
	converted, clientErr := tryConvertOrIncompatible(dt, *val.Get())
	if clientErr != nil {
		return Value(nil), clientErr
	}
	return validateConverted(converted)
}

func validateArrayAfterTryConvert(
	dt FieldDataType,
	val value,
	validateConverted func(value) (value, *ft.ClientErrorItem),
) (value, *ft.ClientErrorItem) {
	if val.Get() == nil {
		return Value(nil), NewInvalidDataTypeErr("")
	}
	rv := reflect.ValueOf(*val.Get())
	if rv.Kind() != reflect.Slice {
		return Value(nil), NewInvalidDataTypeErr("")
	}
	n := rv.Len()
	result := make([]any, n)
	for i := 0; i < n; i++ {
		elem := rv.Index(i).Interface()
		converted, clientErr := tryConvertOrIncompatible(dt, elem)
		if clientErr != nil {
			return Value(nil), clientErr
		}
		validated, clientErr := validateConverted(converted)
		if clientErr != nil {
			return Value(nil), clientErr
		}
		result[i] = *validated.Get()
	}
	return Value(result), nil
}

func toString(value any) (string, error) {
	if value == nil {
		return "", errors.New("toString: value cannot be nil")
	}
	switch v := value.(type) {
	case string:
		return v, nil
	case *string:
		if v == nil {
			return "", errors.New("toString: value cannot be nil")
		}
		return *v, nil
	case fmt.Stringer:
		return v.String(), nil
	default:
		return fmt.Sprint(value), nil
	}
}

func toInt64(value any) (int64, error) {
	unwrapped, err := unwrapOnePointerLevel(value)
	if err != nil {
		return 0, err
	}
	rv := reflect.ValueOf(unwrapped)
	if rv.Kind() == reflect.String {
		return strconv.ParseInt(rv.String(), 10, 64)
	}
	if !rv.Type().ConvertibleTo(reflectTypeInt64) {
		return 0, errors.Errorf("toInt64: cannot convert %T to integer", unwrapped)
	}
	return rv.Convert(reflectTypeInt64).Int(), nil
}

func toInt(value any) (int, error) {
	unwrapped, err := unwrapOnePointerLevel(value)
	if err != nil {
		return 0, err
	}
	rv := reflect.ValueOf(unwrapped)
	if rv.Kind() == reflect.String {
		n, parseErr := strconv.Atoi(rv.String())
		if parseErr != nil {
			return 0, parseErr
		}
		return n, nil
	}
	if !rv.Type().ConvertibleTo(reflectTypeInt) {
		return 0, errors.Errorf("toInt: cannot convert %T to int", unwrapped)
	}
	return int(rv.Convert(reflectTypeInt).Int()), nil
}

func toBool(value any) (bool, error) {
	unwrapped, err := unwrapOnePointerLevel(value)
	if err != nil {
		return false, err
	}
	rv := reflect.ValueOf(unwrapped)
	if rv.Kind() == reflect.String {
		return parseLooseBoolString(rv.String())
	}
	if !rv.Type().ConvertibleTo(reflectTypeBool) {
		return false, errors.Errorf("toBool: cannot convert %T to boolean", unwrapped)
	}
	return rv.Convert(reflectTypeBool).Bool(), nil
}

func toDate(value any) (model.ModelDate, error) {
	if value == nil {
		return model.ModelDate{}, errors.New("toDate: value cannot be nil")
	}
	switch v := value.(type) {
	case time.Time:
		return model.ModelDate(v), nil
	case *time.Time:
		if v == nil {
			return model.ModelDate{}, errors.New("toDate: value cannot be nil")
		}
		return model.ModelDate(*v), nil
	case string:
		return model.ParseModelDate(v)
	default:
		return model.ModelDate{}, errors.Errorf("toDate: cannot convert %T to ModelDate", value)
	}
}

func toTime(value any) (model.ModelTime, error) {
	if value == nil {
		return model.ModelTime{}, errors.New("toTime: value cannot be nil")
	}
	switch v := value.(type) {
	case time.Time:
		return model.ModelTime(v), nil
	case *time.Time:
		if v == nil {
			return model.ModelTime{}, errors.New("toTime: value cannot be nil")
		}
		return model.ModelTime(*v), nil
	case string:
		return model.ParseModelTime(v)
	default:
		return model.ModelTime{}, errors.Errorf("toTime: cannot convert %T to ModelTime", value)
	}
}

func toDateTime(value any) (model.ModelDateTime, error) {
	if value == nil {
		return model.ModelDateTime{}, errors.New("toDateTime: value cannot be nil")
	}
	switch v := value.(type) {
	case time.Time:
		return model.ModelDateTime(v), nil
	case *time.Time:
		if v == nil {
			return model.ModelDateTime{}, errors.New("toDateTime: value cannot be nil")
		}
		return model.ModelDateTime(*v), nil
	case string:
		return model.ParseModelDateTime(v)
	default:
		return model.ModelDateTime{}, errors.Errorf("toDateTime: cannot convert %T to ModelDateTime", value)
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

func getScale(options FieldDataTypeOptions) int {
	if options == nil {
		return -1
	}
	raw, ok := options[FieldDataTypeOptScale]
	if !ok || raw == nil {
		return -1
	}
	switch v := raw.(type) {
	case uint:
		return int(v)
	case uint8:
		return int(v)
	case uint16:
		return int(v)
	case uint32:
		return int(v)
	case uint64:
		return int(v)
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

func toDecimal(value any) (decimal.Decimal, error) {
	switch typed := value.(type) {
	case decimal.Decimal:
		return typed, nil
	case *decimal.Decimal:
		if typed == nil {
			return decimal.Decimal{}, errors.New("toDecimal: value cannot be nil")
		}
		return *typed, nil
	default:
		return decimal.Decimal{}, errors.Errorf("toDecimal: cannot convert %T to decimal", value)
	}
}

func getInt64Range(options FieldDataTypeOptions) []int64 {
	if options == nil {
		return nil
	}
	raw, ok := options[FieldDataTypeOptRange]
	if !ok || raw == nil {
		return nil
	}
	limits, ok := raw.([]int64)
	if !ok || len(limits) != 2 {
		return nil
	}
	return limits
}

func getIntRange(options FieldDataTypeOptions) []int {
	if options == nil {
		return nil
	}
	raw, ok := options[FieldDataTypeOptRange]
	if !ok || raw == nil {
		return nil
	}
	limits, ok := raw.([]int)
	if !ok || len(limits) != 2 {
		return nil
	}
	return limits
}

func getDecimalRange(options FieldDataTypeOptions) ([]decimal.Decimal, error) {
	if options == nil {
		return nil, nil
	}
	raw, ok := options[FieldDataTypeOptRange]
	if !ok || raw == nil {
		return nil, nil
	}
	rangeValues, ok := raw.([]string)
	if !ok || len(rangeValues) != 2 {
		return nil, errors.New("getDecimalRange: invalid range options")
	}
	min, err := decimal.NewFromString(rangeValues[0])
	if err != nil {
		return nil, err
	}
	max, err := decimal.NewFromString(rangeValues[1])
	if err != nil {
		return nil, err
	}
	return []decimal.Decimal{min, max}, nil
}

func applyDecimalScale(value decimal.Decimal, options FieldDataTypeOptions) decimal.Decimal {
	scale := getScale(options)
	if scale < 0 {
		return value
	}
	return value.Round(int32(scale))
}

func validateRangeOrPanic[T ~int | ~int64](min T, max T, fnName string) {
	if min <= max {
		return
	}
	panic(errors.Errorf("%s: min must be less than or equal to max", fnName))
}

func validateDecimalRangeAndScaleOrPanic(min string, max string, scale uint) {
	if scale > 20 {
		panic(errors.New("FieldDataTypeDecimal: scale cannot be greater than 20"))
	}
	minDecimal, err := decimal.NewFromString(min)
	if err != nil {
		panic(errors.Wrap(err, "FieldDataTypeDecimal: invalid min decimal"))
	}
	maxDecimal, err := decimal.NewFromString(max)
	if err != nil {
		panic(errors.Wrap(err, "FieldDataTypeDecimal: invalid max decimal"))
	}
	if minDecimal.GreaterThan(maxDecimal) {
		panic(errors.New("FieldDataTypeDecimal: min must be less than or equal to max"))
	}
}

// unwrapOnePointerLevel dereferences a single *T when value is a non-nil pointer; otherwise
// returns value unchanged. Nil interface or nil pointer returns an error.
func unwrapOnePointerLevel(value any) (any, error) {
	if value == nil {
		return nil, errors.New("unwrapOnePointerLevel: value cannot be nil")
	}
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Ptr {
		return value, nil
	}
	if rv.IsNil() {
		return nil, errors.New("unwrapOnePointerLevel: value cannot be nil")
	}
	return rv.Elem().Interface(), nil
}

func parseLooseBoolString(raw string) (bool, error) {
	s := strings.ToLower(strings.TrimSpace(raw))
	if s == "true" || s == "1" || s == "yes" {
		return true, nil
	}
	if s == "false" || s == "0" || s == "no" {
		return false, nil
	}
	return false, errors.Errorf("parseLooseBoolString: cannot parse '%s' as boolean", raw)
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

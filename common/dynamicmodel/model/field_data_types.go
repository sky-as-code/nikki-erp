package model

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"
	"go.bryk.io/pkg/ulid"

	"github.com/sky-as-code/nikki-erp/common/defense"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	erpjson "github.com/sky-as-code/nikki-erp/common/json"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
)

// FieldDataTypeName is the canonical string returned by FieldDataType.String() and used as a column / generic type id.
const (
	FieldDataTypeNameBoolean       = "boolean"
	FieldDataTypeNameDecimal       = "decimal"
	FieldDataTypeNameEmail         = "email"
	FieldDataTypeNameEnumInt32     = "enumInt32"
	FieldDataTypeNameEnumString    = "enumString"
	FieldDataTypeNameInt32         = "int32"
	FieldDataTypeNameInt64         = "int64"
	FieldDataTypeNameJsonMap       = "jsonmap"
	FieldDataTypeNameModel         = "model"
	FieldDataTypeNameModelDate     = "nikkiDate"
	FieldDataTypeNameModelDateTime = "nikkiDateTime"
	FieldDataTypeNameEtag          = "nikkiEtag"
	FieldDataTypeNameLangCode      = "nikkiLangCode"
	FieldDataTypeNameLangJson      = "nikkiLangJson"
	FieldDataTypeNameSlug          = "nikkiSlug"
	FieldDataTypeNameModelTime     = "nikkiTime"
	FieldDataTypeNamePhone         = "phone"
	FieldDataTypeNameSecret        = "secret"
	FieldDataTypeNameString        = "string"
	FieldDataTypeNameUlid          = "ulid"
	FieldDataTypeNameUrl           = "url"
	FieldDataTypeNameUuid          = "uuid"
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
	return fieldDataTypeEmail{fieldDataTypeBase{name: FieldDataTypeNameEmail, options: nil}}
}

func FieldDataTypePhone() FieldDataType {
	return fieldDataTypePhone{fieldDataTypeBase{name: FieldDataTypeNamePhone, options: nil}}
}

type FieldDataTypeStringOpts struct {
	SanitizeType SanitizeType
	Regex        *regexp.Regexp
}

func FieldDataTypeString(minLength int, maxLength int, stringOpts ...FieldDataTypeStringOpts) FieldDataType {
	validateRangeOrPanic(minLength, maxLength, "FieldDataTypeString")
	dtOpts := processStringOpts(minLength, maxLength, stringOpts...)
	return fieldDataTypeString{fieldDataTypeBase{name: FieldDataTypeNameString, options: dtOpts}}
}

func processStringOpts(minLength int, maxLength int, stringOpts ...FieldDataTypeStringOpts) FieldDataTypeOptions {
	result := FieldDataTypeOptions{
		FieldDataTypeOptLength: []int{minLength, maxLength},
	}
	if len(stringOpts) == 0 {
		return result
	}
	opts := stringOpts[0]
	if opts.SanitizeType != "" {
		result[FieldDataTypeOptSanitizeType] = opts.SanitizeType
	}
	if opts.Regex != nil {
		result[FieldDataTypeOptPattern] = opts.Regex
	}
	return result
}

func FieldDataTypeSecret(minLength int, maxLength int) FieldDataType {
	return fieldDataTypeString{fieldDataTypeBase{
		name: FieldDataTypeNameSecret,
		options: FieldDataTypeOptions{
			FieldDataTypeOptSanitizeType: SanitizeTypeNone,
			FieldDataTypeOptLength:       []int{minLength, maxLength},
		},
	}}
}

func FieldDataTypeUrl() FieldDataType {
	return fieldDataTypeUrl{fieldDataTypeBase{
		name: FieldDataTypeNameUrl,
		options: FieldDataTypeOptions{
			FieldDataTypeOptLength: []int{model.MODEL_RULE_URL_LENGTH_MIN, model.MODEL_RULE_URL_LENGTH_MAX},
		},
	}}
}

func FieldDataTypeUlid() FieldDataType {
	return fieldDataTypeUlid{fieldDataTypeBase{name: FieldDataTypeNameUlid, options: nil}}
}

func FieldDataTypeUuid() FieldDataType {
	return fieldDataTypeUuid{fieldDataTypeBase{name: FieldDataTypeNameUuid, options: nil}}
}

func FieldDataTypeInt64(min int64, max int64) FieldDataType {
	validateRangeOrPanic(min, max, "FieldDataTypeInt64")
	opts := FieldDataTypeOptions{FieldDataTypeOptRange: []int64{min, max}}
	return fieldDataTypeInt64{fieldDataTypeBase{name: FieldDataTypeNameInt64, options: opts}}
}

func FieldDataTypeInt32(min int32, max int32) FieldDataType {
	validateRangeOrPanic(min, max, "FieldDataTypeInt32")
	opts := FieldDataTypeOptions{FieldDataTypeOptRange: []int32{min, max}}
	return fieldDataTypeInt32{fieldDataTypeBase{name: FieldDataTypeNameInt32, options: opts}}
}

func FieldDataTypeDecimal(min string, max string, scale uint) FieldDataType {
	validateDecimalRangeAndScaleOrPanic(min, max, scale)
	opts := FieldDataTypeOptions{
		FieldDataTypeOptRange: []string{min, max},
		FieldDataTypeOptScale: scale,
	}
	return fieldDataTypeDecimal{fieldDataTypeBase{name: FieldDataTypeNameDecimal, options: opts}}
}

func FieldDataTypeBoolean() FieldDataType {
	return fieldDataTypeBoolean{fieldDataTypeBase{name: FieldDataTypeNameBoolean, options: nil}}
}

func FieldDataTypeDate() FieldDataType {
	return fieldDataTypeDate{fieldDataTypeBase{name: FieldDataTypeNameModelDate, options: nil}}
}

func FieldDataTypeTime() FieldDataType {
	return fieldDataTypeTime{fieldDataTypeBase{name: FieldDataTypeNameModelTime, options: nil}}
}

func FieldDataTypeDateTime() FieldDataType {
	return fieldDataTypeDateTime{fieldDataTypeBase{name: FieldDataTypeNameModelDateTime, options: nil}}
}

func FieldDataTypeEnumString(enumValues []string) FieldDataType {
	opts := FieldDataTypeOptions{FieldDataTypeOptEnumValues: enumValues}
	return fieldDataTypeEnumString{fieldDataTypeBase{name: FieldDataTypeNameEnumString, options: opts}}
}

func FieldDataTypeEnumInt32(enumValues []int32) FieldDataType {
	opts := FieldDataTypeOptions{FieldDataTypeOptEnumValues: enumValues}
	return fieldDataTypeEnumInt32{fieldDataTypeBase{name: FieldDataTypeNameEnumInt32, options: opts}}
}

func FieldDataTypeEtag() FieldDataType {
	return fieldDataTypeEtag{fieldDataTypeBase{
		name: FieldDataTypeNameEtag,
		options: FieldDataTypeOptions{
			FieldDataTypeOptLength: []int{model.MODEL_RULE_ETAG_MIN_LENGTH, model.MODEL_RULE_ETAG_MAX_LENGTH},
		},
	}}
}

func FieldDataTypeLangJson(minLength int, maxLength int, stringOpts ...FieldDataTypeStringOpts) FieldDataType {
	validateRangeOrPanic(minLength, maxLength, "FieldDataTypeLangJson")
	dtOpts := processStringOpts(minLength, maxLength, stringOpts...)
	dtOpts[FieldDataTypeOptLangJsonWhitelist] = []model.LanguageCode{
		model.DefaultLanguageCode,
	}
	return fieldDataTypeLangJson{fieldDataTypeBase{name: FieldDataTypeNameLangJson, options: dtOpts}}
}

func FieldDataTypeLangCode() FieldDataType {
	return fieldDataTypeLangCode{fieldDataTypeBase{name: FieldDataTypeNameLangCode, options: nil}}
}

func FieldDataTypeSlug() FieldDataType {
	return fieldDataTypeSlug{fieldDataTypeBase{
		name: FieldDataTypeNameSlug,
		options: FieldDataTypeOptions{
			FieldDataTypeOptLength: []int{model.MODEL_RULE_SLUG_LENGTH_MIN, model.MODEL_RULE_SLUG_LENGTH_MAX},
		},
	}}
}

// FieldDataTypeModel represents a virtual/implicit field that holds a related model or slice of models.
// It is not persisted as a DB column; it is used for graph traversal and API response expansion.
func FieldDataTypeModel() FieldDataType {
	return fieldDataTypeModel{fieldDataTypeBase{name: FieldDataTypeNameModel, options: nil}}
}

func IsFieldDataTypeModel(dt FieldDataType) bool {
	return dt.String() == FieldDataTypeNameModel
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
	if this.isArray {
		return tryConvertStringArrayValue(val)
	}
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
	if this.isArray {
		return tryConvertStringArrayValue(val)
	}
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
	if clientErr := validateStringPattern(out, options); clientErr != nil {
		return Value(nil), clientErr
	}
	return Value(out), nil
}

func validateStringPattern(s string, options FieldDataTypeOptions) *ft.ClientErrorItem {
	if options == nil {
		return nil
	}
	raw, ok := options[FieldDataTypeOptPattern]
	if !ok || raw == nil {
		return nil
	}
	re, ok := raw.(*regexp.Regexp)
	if !ok || re == nil {
		return nil
	}
	if !re.MatchString(s) {
		return ft.NewAnonymousValidationError(
			ft.ErrorKey("err_invalid_string_pattern"),
			"string must match the required pattern",
			nil,
		)
	}
	return nil
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
	if this.isArray {
		return tryConvertStringArrayValue(val)
	}
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
	if this.isArray {
		return tryConvertStringArrayValue(val)
	}
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
		return Value(nil), NewInvalidDataTypeErr("", "ULID")
	}
	if _, err := ulid.Parse(s); err != nil {
		return Value(nil), NewInvalidDataTypeErr("", "ULID")
	}
	return sanitized, nil
}

func (this fieldDataTypeUlid) TryConvert(val any, _ FieldDataTypeOptions) (value, error) {
	if this.isArray {
		return tryConvertStringArrayValue(val)
	}
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
		return Value(nil), NewInvalidDataTypeErr("", "UUID")
	}
	return sanitized, nil
}

func (this fieldDataTypeUuid) TryConvert(val any, _ FieldDataTypeOptions) (value, error) {
	if this.isArray {
		return tryConvertStringArrayValue(val)
	}
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
		return Value(nil), NewInvalidDataTypeErr("", "int64")
	}
	n, ok := (*val.Get()).(int64)
	if !ok {
		return Value(nil), NewInvalidDataTypeErr("", "int64")
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
	if this.isArray {
		return tryConvertInt64ArrayValue(val)
	}
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

type fieldDataTypeInt32 struct{ fieldDataTypeBase }

func (this fieldDataTypeInt32) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeInt32) DefaultValue() value {
	return Value(int32(0))
}

func (this fieldDataTypeInt32) Validate(val value) (value, *ft.ClientErrorItem) {
	if this.isArray {
		return validateArrayAfterTryConvert(this, val, this.validateScalar)
	}
	return validateScalarAfterTryConvert(this, val, this.validateScalar)
}

func (this fieldDataTypeInt32) validateScalar(val value) (value, *ft.ClientErrorItem) {
	if val.Get() == nil {
		return Value(nil), NewInvalidDataTypeErr("", "int32")
	}
	n, ok := (*val.Get()).(int32)
	if !ok {
		return Value(nil), NewInvalidDataTypeErr("", "int32")
	}
	limits := getInt32Range(this.options)
	if len(limits) == 2 && (n < limits[0] || n > limits[1]) {
		return Value(nil), ft.NewAnonymousValidationError(
			ft.ErrorKey("err_invalid_number_range"),
			"value must be between {{.min}} and {{.max}}",
			map[string]any{"min": limits[0], "max": limits[1]},
		)
	}
	return val, nil
}

func (this fieldDataTypeInt32) TryConvert(val any, _ FieldDataTypeOptions) (value, error) {
	if this.isArray {
		return tryConvertInt32ArrayValue(val)
	}
	_, sameType := val.(int32)
	_, ptrSameType := val.(*int32)
	if sameType || ptrSameType {
		return Value(val), nil
	}
	result, err := toInt32(val)
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
		return Value(nil), NewInvalidDataTypeErr("", "decimal")
	}
	n, err := toDecimal(*val.Get())
	if err != nil {
		return Value(nil), NewInvalidDataTypeErr("", "decimal")
	}
	minMax, err := getDecimalRange(this.options)
	if err != nil {
		return Value(nil), NewInvalidDataTypeErr("", "decimal")
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
	if this.isArray {
		return tryConvertDecimalArrayValue(val)
	}
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
	if this.isArray {
		return tryConvertBoolArrayValue(val)
	}
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
	if this.isArray {
		return tryConvertStringArrayValue(val)
	}
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

type fieldDataTypeEnumInt32 struct{ fieldDataTypeBase }

func (this fieldDataTypeEnumInt32) ArrayType() FieldDataType {
	this.isArray = true
	return this
}

func (this fieldDataTypeEnumInt32) DefaultValue() value {
	return Value(int64(0))
}

func (this fieldDataTypeEnumInt32) Validate(value value) (value, *ft.ClientErrorItem) {
	if this.isArray {
		return validateArrayAfterTryConvert(this, value, this.validateScalar)
	}
	return validateScalarAfterTryConvert(this, value, this.validateScalar)
}

func (this fieldDataTypeEnumInt32) validateScalar(value value) (value, *ft.ClientErrorItem) {
	allowed := getEnumNumberValues(this.options)
	if len(allowed) == 0 {
		return value, nil
	}
	allowedAny := make([]any, len(allowed))
	for i, n := range allowed {
		allowedAny[i] = n
	}
	if value.Get() == nil {
		return Value(nil), NewInvalidDataTypeErr("", "int32")
	}
	if err := ValidateOneOf(*value.Get(), allowedAny); err != nil {
		return Value(nil), err
	}
	return value, nil
}

func (this fieldDataTypeEnumInt32) TryConvert(val any, _ FieldDataTypeOptions) (value, error) {
	if this.isArray {
		return tryConvertInt32ArrayValue(val)
	}
	_, sameType := val.(int32)
	_, ptrSameType := val.(*int32)
	if sameType || ptrSameType {
		return Value(val), nil
	}
	result, err := toInt32(val)
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
	if this.isArray {
		return tryConvertStringArrayValue(val)
	}
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

type fieldDataTypeLangJson struct {
	fieldDataTypeBase
}

func (this fieldDataTypeLangJson) ArrayType() FieldDataType {
	panic(errors.New("this field data type does not support array type"))
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
		return Value(nil), NewInvalidDataTypeErr("", "LangJson")
	}
	langObj, clientErr := toLangJson(*value.Get())
	if clientErr != nil {
		return Value(nil), clientErr
	}
	sanitized, _, err := langObj.SanitizeClone(
		getLangJsonWhitelist(this.options),
		this.options[FieldDataTypeOptSanitizeType] == SanitizeTypeHtml,
	)
	for key, val := range *sanitized {
		if cErr := validateStringLength(val, this.options); cErr != nil {
			cErr.Field = "." + key
			return Value(nil), cErr
		}
	}
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
	case map[string]any:
		langJson := model.LangJson{}
		for k, v := range v {
			str, err := toString(v)
			if err != nil {
				return Value(nil), err
			}
			langJson[model.LanguageCode(k)] = str
		}
		return Value(langJson), nil
	case []byte:
		return tryConvertLangJsonBytes(v)
	case string:
		return tryConvertLangJsonBytes([]byte(v))
	default:
		return Value(nil), errors.Errorf(
			"fieldDataTypeLangJson.TryConvert: cannot convert %T to LangJson", val,
		)
	}
}

// FieldDataTypeJsonMap stores arbitrary JSON objects in jsonb columns (map[string]any).
func FieldDataTypeJsonMap() FieldDataType {
	return fieldDataTypeJsonMap{fieldDataTypeBase{name: FieldDataTypeNameJsonMap, options: nil}}
}

type fieldDataTypeJsonMap struct{ fieldDataTypeBase }

func (this fieldDataTypeJsonMap) ArrayType() FieldDataType {
	panic(errors.New("this field data type does not support array type"))
}

func (this fieldDataTypeJsonMap) DefaultValue() value {
	return Value(map[string]any{})
}

func (this fieldDataTypeJsonMap) Validate(val value) (value, *ft.ClientErrorItem) {
	if this.isArray {
		return validateArrayAfterTryConvert(this, val, this.validateScalar)
	}
	return validateScalarAfterTryConvert(this, val, this.validateScalar)
}

func (this fieldDataTypeJsonMap) validateScalar(value value) (value, *ft.ClientErrorItem) {
	if value.Get() == nil {
		return Value(nil), nil
	}
	if _, ok := jsonMapFromAny(*value.Get()); !ok {
		return Value(nil), NewInvalidDataTypeErr("", "JSON Map")
	}
	return value, nil
}

// // fieldDataTypeJsonMap implements MarshallText interface
// func (this fieldDataTypeJsonMap) MarshalText(value value) ([]byte, error) {
// 	if value.Get() == nil {
// 		return nil, nil
// 	}
// 	m, ok := jsonMapFromAny(*value.Get())
// 	if !ok {
// 		return nil, errors.Errorf("fieldDataTypeJsonMap.MarshalText: cannot convert %T to map", value.Get())
// 	}
// 	return json.Marshal(m)
// }

func jsonMapFromAny(v any) (map[string]any, bool) {
	switch m := v.(type) {
	case map[string]any:
		return m, true
	default:
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Map && rv.Type().Key().Kind() == reflect.String {
			out := make(map[string]any, rv.Len())
			iter := rv.MapRange()
			for iter.Next() {
				out[iter.Key().String()] = iter.Value().Interface()
			}
			return out, true
		}
		return nil, false
	}
}

func (this fieldDataTypeJsonMap) TryConvert(val any, _ FieldDataTypeOptions) (value, error) {
	if val == nil {
		return Value(nil), nil
	}
	switch raw := val.(type) {
	case []byte:
		return tryConvertJsonMapBytes(raw)
	case string:
		return tryConvertJsonMapBytes([]byte(raw))
	}
	rv := reflect.ValueOf(val)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return Value(nil), nil
		}
		val = rv.Elem().Interface()
	}
	m, ok := jsonMapFromAny(val)
	if !ok {
		return Value(nil), errors.Errorf("fieldDataTypeJsonMap.TryConvert: cannot convert %T to map", val)
	}
	return Value(m), nil
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
	if this.isArray {
		return tryConvertLangCodeArrayValue(val)
	}
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
		return Value(nil), NewInvalidDataTypeErr("", "slug")
	}
	return sanitized, nil
}

func (this fieldDataTypeSlug) TryConvert(val any, _ FieldDataTypeOptions) (value, error) {
	if this.isArray {
		return tryConvertSlugArrayValue(val)
	}
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
	reflectTypeInt32   = reflect.TypeOf(int32(0))
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
	raw := *val.Get()
	if converted, clientErr := tryConvertOrIncompatible(dt, raw); clientErr == nil {
		if converted.Get() != nil {
			rv := reflect.ValueOf(*converted.Get())
			if rv.Kind() == reflect.Slice {
				n := rv.Len()
				result := make([]any, n)
				for i := 0; i < n; i++ {
					elem := rv.Index(i).Interface()
					validated, elemErr := validateConverted(Value(elem))
					if elemErr != nil {
						return Value(nil), elemErr
					}
					result[i] = *validated.Get()
				}
				return Value(result), nil
			}
		}
	}
	rv := reflect.ValueOf(raw)
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

func toInt32(value any) (int32, error) {
	unwrapped, err := unwrapOnePointerLevel(value)
	if err != nil {
		return 0, err
	}
	rv := reflect.ValueOf(unwrapped)
	if rv.Kind() == reflect.String {
		n, parseErr := strconv.ParseInt(rv.String(), 10, 32)
		if parseErr != nil {
			return 0, parseErr
		}
		return int32(n), nil
	}
	if !rv.Type().ConvertibleTo(reflectTypeInt32) {
		return 0, errors.Errorf("toInt32: cannot convert %T to int32", unwrapped)
	}
	return int32(rv.Convert(reflectTypeInt32).Int()), nil
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
		return model.WrapModelDate(v), nil
	case *time.Time:
		if v == nil {
			return model.ModelDate{}, errors.New("toDate: value cannot be nil")
		}
		return model.WrapModelDate(*v), nil
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
		return model.WrapModelTime(v), nil
	case *time.Time:
		if v == nil {
			return model.ModelTime{}, errors.New("toTime: value cannot be nil")
		}
		return model.WrapModelTime(*v), nil
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
		return model.WrapModelDateTime(v), nil
	case *time.Time:
		if v == nil {
			return model.ModelDateTime{}, errors.New("toDateTime: value cannot be nil")
		}
		return model.WrapModelDateTime(*v), nil
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

func tryConvertStringArrayValue(value any) (value, error) {
	switch typed := value.(type) {
	case []byte:
		return scanPostgresStringArrayBytes(typed)
	case []string:
		return Value(typed), nil
	case *[]string:
		if typed == nil {
			return Value(nil), errors.New("tryConvertStringArrayValue: value cannot be nil")
		}
		return Value(*typed), nil
	default:
		return Value(nil), errors.Errorf("tryConvertStringArrayValue: cannot convert %T to []string", value)
	}
}

func tryConvertInt64ArrayValue(value any) (value, error) {
	switch typed := value.(type) {
	case []byte:
		return scanPostgresInt64ArrayBytes(typed)
	case []int64:
		return Value(typed), nil
	case *[]int64:
		if typed == nil {
			return Value(nil), errors.New("tryConvertInt64ArrayValue: value cannot be nil")
		}
		return Value(*typed), nil
	default:
		return Value(nil), errors.Errorf("tryConvertInt64ArrayValue: cannot convert %T to []int64", value)
	}
}

func tryConvertInt32ArrayValue(value any) (value, error) {
	switch typed := value.(type) {
	case []byte:
		parsed, err := scanPostgresInt32ArrayBytes(typed)
		if err != nil {
			return Value(nil), err
		}
		return Value(parsed), nil
	case []int32:
		return Value(typed), nil
	case *[]int32:
		if typed == nil {
			return Value(nil), errors.New("tryConvertInt32ArrayValue: value cannot be nil")
		}
		return Value(*typed), nil
	default:
		return Value(nil), errors.Errorf("tryConvertInt32ArrayValue: cannot convert %T to []int32", value)
	}
}

func tryConvertBoolArrayValue(value any) (value, error) {
	switch typed := value.(type) {
	case []byte:
		return scanPostgresBoolArrayBytes(typed)
	case []bool:
		return Value(typed), nil
	case *[]bool:
		if typed == nil {
			return Value(nil), errors.New("tryConvertBoolArrayValue: value cannot be nil")
		}
		return Value(*typed), nil
	default:
		return Value(nil), errors.Errorf("tryConvertBoolArrayValue: cannot convert %T to []bool", value)
	}
}

func tryConvertDecimalArrayValue(value any) (value, error) {
	switch typed := value.(type) {
	case []byte:
		return scanPostgresDecimalArrayBytes(typed)
	case []decimal.Decimal:
		return Value(typed), nil
	case *[]decimal.Decimal:
		if typed == nil {
			return Value(nil), errors.New("tryConvertDecimalArrayValue: value cannot be nil")
		}
		return Value(*typed), nil
	default:
		return Value(nil), errors.Errorf("tryConvertDecimalArrayValue: cannot convert %T to []decimal.Decimal", value)
	}
}

func tryConvertLangCodeArrayValue(value any) (value, error) {
	stringsValue, err := tryConvertStringArrayValue(value)
	if err != nil {
		return Value(nil), err
	}
	if stringsValue.Get() == nil {
		return Value(nil), errors.New("tryConvertLangCodeArrayValue: value cannot be nil")
	}
	raw := (*stringsValue.Get()).([]string)
	out := make([]string, len(raw))
	for i := range raw {
		canonical, convErr := model.ToBCP47LanguageCode(raw[i])
		if convErr != nil {
			return Value(nil), convErr
		}
		out[i] = canonical
	}
	return Value(out), nil
}

func tryConvertSlugArrayValue(value any) (value, error) {
	stringsValue, err := tryConvertStringArrayValue(value)
	if err != nil {
		return Value(nil), err
	}
	if stringsValue.Get() == nil {
		return Value(nil), errors.New("tryConvertSlugArrayValue: value cannot be nil")
	}
	raw := (*stringsValue.Get()).([]string)
	out := make([]string, len(raw))
	for i := range raw {
		item := strings.ToLower(strings.TrimSpace(raw[i]))
		out[i] = strings.ReplaceAll(item, " ", "-")
	}
	return Value(out), nil
}

func scanPostgresStringArrayBytes(raw []byte) (value, error) {
	var arr pq.StringArray
	if err := arr.Scan(raw); err != nil {
		return Value(nil), err
	}
	return Value([]string(arr)), nil
}

func scanPostgresBoolArrayBytes(raw []byte) (value, error) {
	var arr pq.BoolArray
	if err := arr.Scan(raw); err != nil {
		return Value(nil), err
	}
	return Value([]bool(arr)), nil
}

func scanPostgresInt32ArrayBytes(raw []byte) ([]int32, error) {
	var arr pq.Int32Array
	if err := arr.Scan(raw); err != nil {
		return nil, err
	}
	return []int32(arr), nil
}

func scanPostgresInt64ArrayBytes(raw []byte) (value, error) {
	var arr pq.Int64Array
	if err := arr.Scan(raw); err != nil {
		return Value(nil), err
	}
	return Value([]int64(arr)), nil
}

func scanPostgresDecimalArrayBytes(raw []byte) (value, error) {
	var arr pq.StringArray
	if err := arr.Scan(raw); err != nil {
		return Value(nil), err
	}
	out := make([]decimal.Decimal, len(arr))
	for i := range arr {
		item, convErr := decimal.NewFromString(arr[i])
		if convErr != nil {
			return Value(nil), convErr
		}
		out[i] = item
	}
	return Value(out), nil
}

func tryConvertLangJsonBytes(raw []byte) (value, error) {
	if len(raw) == 0 {
		return Value(nil), errors.New("fieldDataTypeLangJson.TryConvert: value cannot be empty")
	}
	var mapped map[string]string
	if err := erpjson.Unmarshal(raw, &mapped); err != nil {
		return Value(nil), err
	}
	return Value(model.LangJson(mapped)), nil
}

func tryConvertJsonMapBytes(raw []byte) (value, error) {
	if len(raw) == 0 {
		return Value(nil), nil
	}
	var mapped map[string]any
	if err := erpjson.Unmarshal(raw, &mapped); err != nil {
		return Value(nil), err
	}
	return Value(mapped), nil
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

func getInt32Range(options FieldDataTypeOptions) []int32 {
	if options == nil {
		return nil
	}
	raw, ok := options[FieldDataTypeOptRange]
	if !ok || raw == nil {
		return nil
	}
	limits, ok := raw.([]int32)
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

func validateRangeOrPanic[T ~int | ~int32 | ~int64](min T, max T, fnName string) {
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

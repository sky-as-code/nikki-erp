package model

import (
	"time"

	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/model"
)

func (this DynamicFields) GetBool(key string) *bool {
	val, ok := this[key]
	if !ok || val == nil {
		return nil
	}
	b := val.(bool)
	return &b
}

func (this DynamicFields) SetBool(key string, v *bool) {
	if v == nil {
		this[key] = nil
		return
	}
	this[key] = *v
}

// GetDecimal reads a decimal field in whatever shape it arrived in.
//
// It used to bare type-assert to decimal.Decimal, which PANICKED on every other shape - and a
// decimal crosses JSON as a STRING precisely so it does not lose precision, so any value read back
// through a jsonb column crashed the caller rather than answering. Four modules had independently
// written their own switch to route around this (accounting/domain/services/values.go,
// inventory/domain/services/stock_scrap_domservice.go, paymentinvoice/dynamicengines/order_actions.go,
// purchase/domain/services/purchase_service.go); accepting the shapes here means they did not have to.
//
// A value that cannot be read as a decimal returns nil, the same answer an absent field gives. That
// is deliberate: the alternative is a panic, and a caller that already handles "not set" handles
// this correctly, while none of them is prepared to recover from a crash mid-transaction.
func (this DynamicFields) GetDecimal(key string) *decimal.Decimal {
	val, ok := this[key]
	if !ok || val == nil {
		return nil
	}

	switch typed := val.(type) {
	case decimal.Decimal:
		return &typed
	case *decimal.Decimal:
		return typed
	case string:
		parsed, err := decimal.NewFromString(typed)
		if err != nil {
			return nil
		}
		return &parsed
	case float64:
		parsed := decimal.NewFromFloat(typed)
		return &parsed
	case float32:
		parsed := decimal.NewFromFloat32(typed)
		return &parsed
	case int:
		parsed := decimal.NewFromInt(int64(typed))
		return &parsed
	case int32:
		parsed := decimal.NewFromInt(int64(typed))
		return &parsed
	case int64:
		parsed := decimal.NewFromInt(typed)
		return &parsed
	}
	return nil
}

func (this DynamicFields) SetDecimalStr(key string, v *string) {
	if v == nil {
		this[key] = nil
		return
	}
	dec, err := decimal.NewFromString(*v)
	if err != nil {
		panic(errors.Wrap(err, "SetDecimalStr"))
	}
	this[key] = dec
}

func (this DynamicFields) SetDecimal(key string, v *decimal.Decimal) {
	if v == nil {
		this[key] = nil
		return
	}
	this[key] = *v
}

// GetString returns the string value at key. Returns nil if key is missing or value is nil.
// Caller must ensure the map is initialized (non-nil).
func (this DynamicFields) GetString(key string) *string {
	val, ok := this[key]
	if !ok || val == nil {
		return nil
	}
	s := val.(string)
	return &s
}

// SetString sets the string value at key.
// Caller must ensure the map is initialized (non-nil).
func (this DynamicFields) SetString(key string, v *string) {
	if v == nil {
		this[key] = nil
		return
	}
	this[key] = *v
}

func (this DynamicFields) GetStrings(key string) []string {
	val, ok := this[key]
	if !ok || val == nil {
		return nil
	}
	s := val.([]string)
	return s
}

func (this DynamicFields) SetStrings(key string, v []string) {
	this[key] = v
}

func (this DynamicFields) GetAny(key string) any {
	val, ok := this[key]
	if !ok || val == nil {
		return nil
	}
	s := val
	return s
}

func (this DynamicFields) SetAny(key string, v any) {
	this[key] = v
}

// GetModelId returns the model.Id value at key. Returns nil if key is missing or value is nil.
// Caller must ensure the map is initialized (non-nil).
func (this DynamicFields) GetModelId(key string) *model.Id {
	val, ok := this[key]
	if !ok || val == nil {
		return nil
	}
	s := val.(string)
	id := model.Id(s)
	return &id
}

// SetModelId sets the model.Id value at key.
// Caller must ensure the map is initialized (non-nil).
func (this DynamicFields) SetModelId(key string, v *model.Id) {
	if v == nil {
		this[key] = nil
		return
	}
	this[key] = string(*v)
}

func (this DynamicFields) GetInt32(key string) *int32 {
	val, ok := this[key]
	if !ok || val == nil {
		return nil
	}
	s := val.(int32)
	return &s
}

func (this DynamicFields) SetInt32(key string, v *int32) {
	if v == nil {
		this[key] = nil
		return
	}
	this[key] = *v
}

// GetInt64 returns the int64 value at key. Returns nil if key is missing or value is nil.
// Caller must ensure the map is initialized (non-nil).
func (this DynamicFields) GetInt64(key string) *int64 {
	val, ok := this[key]
	if !ok || val == nil {
		return nil
	}
	s := val.(int64)
	return &s
}

func (this DynamicFields) SetInt64(key string, v *int64) {
	if v == nil {
		this[key] = nil
		return
	}
	this[key] = *v
}

func (this DynamicFields) MustGetInt64(key string) (result int64) {
	val, ok := this[key]
	if !ok || val == nil {
		return
	}
	return val.(int64)
}

func (this DynamicFields) GetEtag(key string) *model.Etag {
	val, ok := this[key]
	if !ok || val == nil {
		return nil
	}
	etag := val.(model.Etag)
	return &etag
}

func (this DynamicFields) SetEtag(key string, v *model.Etag) {
	if v == nil {
		this[key] = nil
		return
	}
	this[key] = *v
}

func (this DynamicFields) GetModelDateTime(key string) *model.ModelDateTime {
	val, ok := this[key]
	if !ok || val == nil {
		return nil
	}
	var modelDateTime model.ModelDateTime
	switch v := val.(type) {
	case model.ModelDateTime:
		modelDateTime = v
	case time.Time:
		modelDateTime = model.WrapModelDateTime(v)
	default:
		return nil
	}
	return &modelDateTime
}

func (this DynamicFields) SetModelDateTime(key string, v *model.ModelDateTime) {
	if v == nil {
		this[key] = nil
		return
	}
	this[key] = *v
}

func (this DynamicFields) GetModelDate(key string) *model.ModelDate {
	val, ok := this[key]
	if !ok || val == nil {
		return nil
	}
	var modelDate model.ModelDate
	switch v := val.(type) {
	case model.ModelDate:
		modelDate = v
	case time.Time:
		modelDate = model.WrapModelDate(v)
	default:
		return nil
	}
	return &modelDate
}

func (this DynamicFields) SetModelDate(key string, v *model.ModelDate) {
	if v == nil {
		this[key] = nil
		return
	}
	this[key] = *v
}

func (this DynamicFields) GetModelTime(key string) *model.ModelTime {
	val, ok := this[key]
	if !ok || val == nil {
		return nil
	}
	var modelTime model.ModelTime
	switch v := val.(type) {
	case model.ModelTime:
		modelTime = v
	case time.Time:
		modelTime = model.WrapModelTime(v)
	default:
		return nil
	}
	return &modelTime
}

func (this DynamicFields) SetModelTime(key string, v *model.ModelTime) {
	if v == nil {
		this[key] = nil
		return
	}
	this[key] = *v
}

func (this DynamicFields) GetSlug(key string) *model.Slug {
	val, ok := this[key]
	if !ok || val == nil {
		return nil
	}
	slug := val.(model.Slug)
	return &slug
}

func (this DynamicFields) SetSlug(key string, v *model.Slug) {
	if v == nil {
		this[key] = nil
		return
	}
	this[key] = string(*v)
}

// GetLangJson safely extracts a LangJson from DynamicFields.
// Returns nil if key is missing, value is nil, or type assertion fails.
// This method validates each entry to ensure type safety.
func (this DynamicFields) GetLangJson(key string) *model.LangJson {
	val, ok := this[key]
	if !ok || val == nil {
		return nil
	}

	// Try direct type assertion first (when already LangJson)
	if lj, ok := val.(model.LangJson); ok {
		return &lj
	}

	// Handle map[string]interface{} case (from JSON unmarshaling)
	rawMap, ok := val.(map[string]interface{})
	if !ok {
		return nil
	}

	lang := make(model.LangJson)
	for k, v := range rawMap {
		strVal, ok := v.(string)
		if !ok {
			continue
		}
		lang[model.LanguageCode(k)] = strVal
	}

	return &lang
}

// SetLangJson stores a LangJson value in DynamicFields.
// Sets nil if the input value is nil.
func (this DynamicFields) SetLangJson(key string, v *model.LangJson) {
	if v == nil {
		this[key] = nil
		return
	}
	this[key] = *v
}

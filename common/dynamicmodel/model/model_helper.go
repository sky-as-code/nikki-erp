package model

import (
	"time"

	"github.com/shopspring/decimal"

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

func (this DynamicFields) GetDecimal(key string) *decimal.Decimal {
	val, ok := this[key]
	if !ok || val == nil {
		return nil
	}
	d := val.(decimal.Decimal)
	return &d
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

package schema

import (
	"github.com/sky-as-code/nikki-erp/common/model"
)

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
	return val.(*model.Etag)
}

func (this DynamicFields) SetEtag(key string, v *model.Etag) {
	if v == nil {
		this[key] = nil
		return
	}
	this[key] = *v
}

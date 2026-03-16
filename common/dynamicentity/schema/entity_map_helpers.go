package schema

import (
	"github.com/sky-as-code/nikki-erp/common/model"
)

// GetString returns the string value at key. Returns nil if key is missing or value is nil.
// Caller must ensure the map is initialized (non-nil).
func (this DynamicEntity) GetString(key string) *string {
	val, ok := this[key]
	if !ok || val == nil {
		return nil
	}
	s := val.(string)
	return &s
}

// SetString sets the string value at key.
// Caller must ensure the map is initialized (non-nil).
func (this DynamicEntity) SetString(key string, v *string) {
	if v == nil {
		this[key] = nil
		return
	}
	this[key] = *v
}

// GetModelId returns the model.Id value at key. Returns nil if key is missing or value is nil.
// Caller must ensure the map is initialized (non-nil).
func (this DynamicEntity) GetModelId(key string) *model.Id {
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
func (this DynamicEntity) SetModelId(key string, v *model.Id) {
	if v == nil {
		this[key] = nil
		return
	}
	this[key] = string(*v)
}

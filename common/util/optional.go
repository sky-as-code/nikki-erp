package util

import (
	"bytes"
	"encoding/json"
)

// Optional distinguishes three JSON states for the same field:
//
//   - Absent  : key is missing from the JSON body. IsSet() == false.
//   - Null    : key is present with value `null`.   IsSet() == true, Get() == nil.
//   - Value   : key is present with a real value.   IsSet() == true, Get() != nil.
//
// Use it in DTOs when partial-update semantics are required so the mapper
// can skip the field entirely on absence, clear it on null, or set it on value.
type Optional[T any] struct {
	set   bool
	value *T
}

func NewOptional[T any](v T) Optional[T] {
	return Optional[T]{set: true, value: &v}
}

func NewOptionalNull[T any]() Optional[T] {
	return Optional[T]{set: true}
}

func (this Optional[T]) IsSet() bool {
	return this.set
}

func (this Optional[T]) IsNull() bool {
	return this.set && this.value == nil
}

func (this Optional[T]) Get() *T {
	return this.value
}

func (this *Optional[T]) UnmarshalJSON(data []byte) error {
	this.set = true
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		this.value = nil
		return nil
	}
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		this.set = false
		return err
	}
	this.value = &v
	return nil
}

func (this Optional[T]) MarshalJSON() ([]byte, error) {
	if !this.set || this.value == nil {
		return []byte("null"), nil
	}
	return json.Marshal(*this.value)
}

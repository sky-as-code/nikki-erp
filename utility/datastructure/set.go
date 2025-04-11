package datastructure

import (
	"fmt"

	"github.com/sky-as-code/nikki-erp/utility/json"
)

// Set uses map as set of items.
type Set[T comparable] map[T]struct{}

// ToSlice converts the set to a slice.
func (this Set[T]) ToSlice() []T {
	keys := make([]T, 0, len(this))
	for k := range this {
		keys = append(keys, k)
	}
	return keys
}

// IsEmpty checks whether the set is empty or not.
func (this Set[T]) IsEmpty() bool {
	return len(this) == 0
}

// Add adds item to the set.
func (this Set[T]) Add(item T) {
	this[item] = struct{}{}
}

// Remove removes item from the set. It does nothing if item does not exist in the set.
func (this Set[T]) Remove(item T) {
	delete(this, item)
}

// Contains checks if item is in the set.
func (this Set[T]) Contains(item T) bool {
	_, ok := this[item]
	return ok
}

// Length returns number of items in the set.
func (this Set[T]) Length() int {
	return len(this)
}

// String returns printable string of the set.
func (this Set[T]) String() string {
	return fmt.Sprintf("%s", this.ToSlice())
}

// MarshalJSON belongs to json.Marshaler interface.
// It converts Golang object to JSON data.
func (this Set[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(this.ToSlice())
}

// UnmarshalJSON belongs to json.Unmarshaler interface.
// It parses JSON data to Golang object.
func (this *Set[T]) UnmarshalJSON(data []byte) error {
	var err error
	if err = this.unmarshalArray(data); err != nil {
		err = this.unmarshalSingleItem(data)
	}

	return err
}

func (this *Set[T]) unmarshalArray(data []byte) error {
	array := []T{}
	var err error
	if err = json.Unmarshal(data, &array); err == nil {
		*this = make(Set[T])
		for _, item := range array {
			this.Add(item)
		}
	}

	return err
}

func (this *Set[T]) unmarshalSingleItem(data []byte) error {
	var item T
	var err error
	if err = json.Unmarshal(data, &item); err == nil {
		*this = make(Set[T])
		this.Add(item)
	}

	return err
}

// NewSet creates new set.
func NewSet[T comparable]() Set[T] {
	return make(Set[T])
}

// CreateSet creates new set with given values.
func NewSetFrom[T comparable](items ...T) Set[T] {
	set := NewSet[T]()
	for _, k := range items {
		set.Add(k)
	}
	return set
}

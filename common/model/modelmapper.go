package model

import (
	"github.com/go-sanitize/sanitize"

	. "github.com/sky-as-code/nikki-erp/common/fault"
	"gopkg.in/jeevatkm/go-model.v1"
)

func AddConversion[TIn any, TOut any](converter model.Converter) {
	model.AddConversion((*TIn)(nil), (*TOut)(nil), converter)
}

func Copy(dest, src interface{}) error {
	if src == nil {
		return NewTechnicalError("modelmapper.Copy() src is a nil pointer")
	}
	errors := model.Copy(dest, src)
	return WrapTechnicalError(JoinErrors(errors), "modelmapper.Copy() failed")
}

// Clone returns a deep clone of given object
func Clone[T interface{}](src T) (T, error) {
	clone, err := model.Clone(src)
	return clone.(T), WrapTechnicalError(err, "modelmapper.Clone[%T]() failed", src)
}

// ToMap deeply converts a struct into a map[string]any
func ToMap(src any) (map[string]any, error) {
	outputMap, err := model.Map(src)
	return outputMap, WrapTechnicalError(err, "modelmapper.ToMap() failed")
}

var sanitizer, _ = sanitize.New()

func Sanitize(target any) {
	sanitizer.Sanitize(target)
}

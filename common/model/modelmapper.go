package model

import (
	goerrors "errors"
	"fmt"

	"github.com/go-sanitize/sanitize"
	"go.bryk.io/pkg/errors"

	"gopkg.in/jeevatkm/go-model.v1"
)

func AddConversion[TIn any, TOut any](converter model.Converter) {
	model.AddConversion((*TIn)(nil), (*TOut)(nil), converter)
}

func Copy(dest, src interface{}) error {
	if src == nil {
		return errors.New("modelmapper.Copy() src is a nil pointer")
	}
	errs := model.Copy(dest, src)
	return goerrors.Join(errs...)
}

// Clone returns a deep clone of given object
func Clone[T interface{}](src T) (T, error) {
	clone, err := model.Clone(src)
	return clone.(T), errors.Wrap(err, fmt.Sprintf("modelmapper.Clone[%T]() failed", src))
}

// ToMap deeply converts a struct into a map[string]any
func ToMap(src any) (map[string]any, error) {
	outputMap, err := model.Map(src)
	return outputMap, errors.Wrap(err, "modelmapper.ToMap() failed")
}

var sanitizer, _ = sanitize.New()

func Sanitize(target any) {
	sanitizer.Sanitize(target)
}

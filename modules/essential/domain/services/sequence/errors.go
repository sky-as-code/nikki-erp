// Package sequence holds the pure algorithms behind Sequence: the pseudorandom (FF1) value, template rendering and
// validation, reset periods and the idempotency fingerprint. It has no database or module dependency.
package sequence

import (
	"fmt"

	"go.bryk.io/pkg/errors"
)

const (
	CodeCapacityExceeded = "SEQUENCE_CAPACITY_EXCEEDED"
	CodeInvalidTemplate  = "INVALID_TEMPLATE"
)

var (
	// ErrCapacityExceeded is returned once the counter has used every length domain up to MaxLength.
	ErrCapacityExceeded = errors.New(CodeCapacityExceeded)
	// ErrInvalidTemplate matches (errors.Is) every *TemplateError.
	ErrInvalidTemplate = errors.New(CodeInvalidTemplate)
)

// TemplateError is a client error about a prefix/suffix template. Its message starts with INVALID_TEMPLATE.
type TemplateError struct {
	Detail string
}

func (this *TemplateError) Error() string {
	return CodeInvalidTemplate + ": " + this.Detail
}

func (this *TemplateError) Is(target error) bool {
	return target == ErrInvalidTemplate
}

func templateErrorf(format string, args ...any) error {
	return &TemplateError{Detail: fmt.Sprintf(format, args...)}
}

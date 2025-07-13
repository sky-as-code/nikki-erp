package fault

import (
	"fmt"

	invopop "github.com/invopop/validation"
)

type ClientError struct {
	Code    string `json:"code"`
	Details any    `json:"details"`
}

func (this ClientError) Error() string {
	return fmt.Sprintf("%s: %v", this.Code, this.Details)
}

type ValidationErrorItem struct {
	Field string
	Error string
}

type ValidationErrors map[string]string

func (this *ValidationErrors) Append(field string, err string) {
	(*this)[field] = err
}

func (this *ValidationErrors) Appendf(field string, err string, args ...any) {
	(*this)[field] = fmt.Sprintf(err, args...)
}

func (this *ValidationErrors) AppendItem(item ValidationErrorItem) {
	(*this)[item.Field] = item.Error
}

func (this *ValidationErrors) Count() int {
	return len(*this)
}

func (this *ValidationErrors) Has(field string) bool {
	_, ok := (*this)[field]
	return ok
}

func (this *ValidationErrors) Merge(other ValidationErrors) {
	for field, err := range other {
		(*this)[field] = err
	}
}

func (this *ValidationErrors) MergeClientError(other *ClientError) {
	if other != nil {
		otherErrs, isOk := other.Details.(ValidationErrors)
		if isOk {
			this.Merge(otherErrs)
		}
	}
}

func (this *ValidationErrors) Error() string {
	str := ""
	for field, err := range *this {
		str += fmt.Sprintf("%s: %s;", field, err)
	}
	return str
}

func (this *ValidationErrors) ToClientError() *ClientError {
	return &ClientError{
		Code:    "validation_error",
		Details: this,
	}
}

func NewValidationErrors() ValidationErrors {
	return make(ValidationErrors, 0)
}

func NewValidationErrorsFromInvopop(rawErrors invopop.Errors) ValidationErrors {
	errors := make(ValidationErrors, len(rawErrors))
	for field, err := range rawErrors {
		invoErr, ok := err.(invopop.ErrorObject)
		if ok {
			errors.Append(field, invoErr.Error())
			continue
		}
		errors.Append(field, err.Error())
	}
	return errors
}

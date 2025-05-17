package fault

import (
	"fmt"

	invopop "github.com/invopop/validation"
	"github.com/sky-as-code/nikki-erp/common/util"
)

func WrapValidationErrors(err ValidationErrors) *ClientError {
	return &ClientError{
		Code:    "validation_error",
		Details: err,
	}
}

type ClientError struct {
	Code    string `json:"code"`
	Details any    `json:"details"`
}

type ValidationErrorItem struct {
	Field string
	Error string
}

type ValidationErrors map[string]string

func (this *ValidationErrors) Append(item ValidationErrorItem) {
	(*this)[item.Field] = item.Error
}

func (this ValidationErrors) Count() int {
	return len(this)
}

func (this ValidationErrors) Has(field string) bool {
	_, ok := this[field]
	return ok
}

func (this ValidationErrors) Error() string {
	str := ""
	for field, err := range this {
		str += fmt.Sprintf("%s: %s;", field, err)
	}
	return str
}

func NewValidationErrors() ValidationErrors {
	return make(ValidationErrors, 0)
}

func NewValidationErrorsFromInvopop(rawErrors invopop.Errors) ValidationErrors {
	errors := make(ValidationErrors, len(rawErrors))
	util.Unused(invopop.ErrorTag)
	for field, err := range rawErrors {
		e := err.(invopop.ErrorObject)
		errors.Append(ValidationErrorItem{
			Field: field,
			Error: e.Error(),
		})
	}
	return errors
}

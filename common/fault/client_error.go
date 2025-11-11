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

func (this *ValidationErrors) AppendAlreadyExists(fieldName string, fieldLabel string) {
	this.Appendf(fieldName, "%s already exists", fieldLabel)
}

func (this *ValidationErrors) AppendNotFound(fieldName string, fieldLabel string) {
	this.Appendf(fieldName, "%s not found", fieldLabel)
}

func (this *ValidationErrors) AppendConstraintViolated(fieldName string, fieldLabel string) {
	this.Appendf(fieldName, "%s constraint violated", fieldLabel)
}

func (this *ValidationErrors) AppendNotAllowed(fieldName string, fieldLabel string) {
	this.Appendf(fieldName, "%s not allowed", fieldLabel)
}

func (this *ValidationErrors) AppendEtagMismatched() {
	this.Append("etag", "etag mismatched")
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

func (this *ValidationErrors) MergeClientError(other *ClientError) bool {
	if other == nil {
		return true
	}
	otherErrs, isOk := other.Details.(ValidationErrors)
	if isOk {
		for field, err := range otherErrs {
			(*this)[field] = fmt.Sprint(err)
		}
		return true
	}
	return false
}

func (this *ValidationErrors) RenameKey(oldKey string, newKey string) bool {
	val, ok := (*this)[oldKey]
	if !ok {
		return false
	}
	(*this)[newKey] = val
	delete(*this, oldKey)
	return true
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
		Details: *this,
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

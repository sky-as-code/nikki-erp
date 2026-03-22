package fault

import (
	"bytes"
	"fmt"
	"text/template"

	invopop "github.com/invopop/validation"
)

type ValidationErrorItem struct {
	Field   string
	Error   string
	Key     string
	Message string
	Vars    map[string]any
}

// String returns the message with variables substituted.
func (this *ValidationErrorItem) String() string {
	if this == nil || this.Message == "" {
		return ""
	}
	if len(this.Vars) == 0 {
		return this.Message
	}
	tmpl, err := template.New("validation").Parse(this.Message)
	if err != nil {
		return this.Message
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, this.Vars); err != nil {
		return this.Message
	}
	return buf.String()
}

type ClientError struct {
	Code    string `json:"code"`
	Details any    `json:"details"`
}

func (this ClientError) Error() string {
	return fmt.Sprintf("%s: %v", this.Code, this.Details)
}

func NewClientErrors() *ClientErrors {
	return &ClientErrors{}
}

type ClientErrors []ClientErrorItem

func (this *ClientErrors) Append(item ClientErrorItem) {
	*this = append(*this, item)
}

func (this *ClientErrors) Count() int {
	return len(*this)
}

func (this *ClientErrors) Has(field string) bool {
	for _, item := range *this {
		if item.Field == field {
			return true
		}
	}
	return false
}

type ClientErrorType string

const (
	// Error caused by invalid input data, e.g: validation error, missing required fields, etc.
	ClientErrorTypeValidation ClientErrorType = "validation"

	// Error caused by business logic, aka business invariant, violations, e.g: insufficient balance, resource not found, etc.
	ClientErrorTypeBusiness ClientErrorType = "business"
)

func NewBusinessViolation(field string, key string, message string, vars ...map[string]any) *ClientErrorItem {
	var msgVars map[string]any = nil
	if len(vars) == 0 {
		msgVars = vars[0]
	}

	return &ClientErrorItem{
		Field:   field,
		Key:     key,
		Message: message,
		Vars:    msgVars,
		Type:    ClientErrorTypeBusiness,
	}
}

func NewValidationError(field string, key string, message string, vars ...map[string]any) *ClientErrorItem {
	var msgVars map[string]any = nil
	if len(vars) > 0 {
		msgVars = vars[0]
	}

	return &ClientErrorItem{
		Field:   field,
		Key:     key,
		Message: message,
		Vars:    msgVars,
		Type:    ClientErrorTypeValidation,
	}
}

func NewAnonymousValidationError(key string, message string, vars ...map[string]any) *ClientErrorItem {
	var msgVars map[string]any = nil
	if len(vars) > 0 {
		msgVars = vars[0]
	}

	return &ClientErrorItem{
		Field:   "",
		Key:     key,
		Message: message,
		Vars:    msgVars,
		Type:    ClientErrorTypeValidation,
	}
}

type ClientErrorItem struct {
	// Field name in request payload that caused the error
	Field string `json:"field,omitempty"`

	// Translation key
	Key string `json:"key,omitempty"`

	// Error message template, support variable substitution.
	// This is for human-friendly error logging.
	// To display to end user, use `Key` and optional `Vars` to localize the error message.
	Message string `json:"message,omitempty"`

	// Error type. Can be used to determine the position of the error message on the UI,
	// or used for analytics.
	Type ClientErrorType `json:"type,omitempty"`

	// Variables for substitution into Message.
	Vars map[string]any `json:"vars,omitempty"`
}

// String returns the message with variables substituted.
func (this ClientErrorItem) String() string {
	if this.Message == "" {
		return ""
	}
	if len(this.Vars) == 0 {
		return this.Message
	}
	tmpl, err := template.New("ClientErrorItem").Parse(this.Message)
	if err != nil {
		return this.Message
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, this.Vars); err != nil {
		return this.Message
	}
	if this.Key != "" {
		return fmt.Sprintf("%s: %s", this.Field, buf.String())
	}
	return buf.String()
}

type ValidationErrorCollection map[string]string

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

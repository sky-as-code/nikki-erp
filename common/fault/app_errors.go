package fault

import (
	"fmt"
	"strings"

	apiVal "github.com/invopop/validation"
)

// type error interface {
// 	error
// 	IsTechnical() bool
// 	// Message returns this error's own message.
// 	// Meanwhile Error() returns the concatenation of
// 	// this error's and internal's messages.
// 	Message() string
// 	DisplayMessage() string
// 	SetDisplayMessage(message string, formatValues ...interface{})
// 	Name() string
// 	// Unwrap is for supporting the errors.Unwrap(err) since Go v1.13
// 	Unwrap() error
// 	GetError() error
// 	GetErrorCode() string
// }

type errorBase struct {
	internal error
	message  string
	// For errors that want to be displayed on the frontend, they are not the same as the actual message.
	//Especially technical errors are often displayed differently from the message
	displayMessage string
	name           string
	errorCode      string
}

// Error() returns the concatenation of this error's and internal's messages.
// If you want to return simple error message to client, use Message() function.
func (err errorBase) Error() string {
	if err.internal != nil {
		if len(err.Message()) > 0 {
			return fmt.Sprintf("%d: %s: %s", err.errorCode, err.Message(), err.internal.Error())
		}
		return err.internal.Error()
	}
	return err.Message()
}

func (err errorBase) GetError() error {
	return err.internal
}

func (err errorBase) GetErrorCode() string {
	return err.errorCode
}

func (err errorBase) Name() string {
	return err.name
}

func (err errorBase) Unwrap() error {
	return err.internal
}

func (errorBase) IsTechnical() bool {
	return false
}

// Message() returns this error's own message.
// Meanwhile Error() returns the concatenation of
// this error's and internal's messages.
func (err errorBase) Message() string {
	return err.message
}

func (err errorBase) SetDisplayMessage(message string, formatValues ...interface{}) {
	err.displayMessage = fmt.Sprintf(message, formatValues...)
}

func (err errorBase) DisplayMessage() string {
	return err.displayMessage
}

// TechnicalError represents an error caused by developers and should be logged
type TechnicalError struct {
	errorBase
}

func (TechnicalError) IsTechnical() bool {
	return true
}

func (TechnicalError) Name() string {
	return "TechnicalError"
}

// NewTechnicalError creates an instance of TechnicalError.
// Internally it uses fmt.Sprintf to format the message with formatValues.
//
// NewTechnicalError("incorrect database credentials")
//
// NewTechnicalError("insufficient permission to read file name %s", fileName)
func NewTechnicalError(message string, formatValues ...interface{}) TechnicalError {
	base := errorBase{
		message: fmt.Sprintf(message, formatValues...),
	}
	err := TechnicalError{
		errorBase: base,
	}
	return err
}

// WrapTechnicalError creates an instance of TechnicalError which in turn wraps
// the provided err. Internally it uses fmt.Sprintf to format the message with formatValues.
//
// data, err := FailToReadFile(userId)
// techErr := WrapTechnicalError(err, "loadUserList")
//
// Calling techErr.Error() returns "loadUserList: insufficient permission to read file"
//
// data, err := FailToReadFile(userId)
// techErr := WrapTechnicalError(err, "loadUserList(%s)", userId)
//
// Calling techErr.Error() returns "loadUserList(123): insufficient permission to read file"
func WrapTechnicalError(err error, message string, formatValues ...interface{}) error {
	if err == nil {
		return nil
	}

	base := errorBase{
		internal: err,
		message:  fmt.Sprintf(message, formatValues...),
	}
	return TechnicalError{
		errorBase: base,
	}
}

func WrapError(errCode string, err error, message string, formatValues ...interface{}) error {
	if err == nil {
		return nil
	}

	base := errorBase{
		errorCode: errCode,
		internal:  err,
		message:   fmt.Sprintf(message, formatValues...),
	}
	return TechnicalError{
		errorBase: base,
	}
}

// BusinessError represents an error caused by users providing data that
// violate business rules. This kind of error should result in a meaningful
// response to users and doesn't need logging.
type BusinessError struct {
	errorBase
}

func (BusinessError) Name() string {
	return "BusinessError"
}

// NewBusinessError creates an instance of BusinessError.
// Internally it uses fmt.Sprintf to format the message with formatValues.
//
// NewBusinessError("cart is full")
//
// NewBusinessError("user %s has been banned", user.Name())
func NewBusinessError(errorCode string, message string, formatValues ...interface{}) BusinessError {
	base := errorBase{
		errorCode: errorCode,
		message:   fmt.Sprintf(message, formatValues...),
	}
	return BusinessError{
		errorBase: base,
	}
}

// WrapBusinessError creates an instance of BusinessError which in turn wraps
// the provided err. Internally it uses fmt.Sprintf to format the message with formatValues.
//
// _, err := AddToCart(user, product)
// techErr := WrapBusinessError(err, "buyProduct")
//
// Calling techErr.Error() returns "buyProduct: cart is full"
//
// _, err := AddToCart(user, product)
// techErr := WrapBusinessError(err, "buyProduct(%s, %s)", user.Name(), product.Name())
//
// Calling techErr.Error() returns "buyProduct(Adam, Keyboard): Cart is full"
func WrapBusinessError(err error, message string, formatValues ...interface{}) error {
	if err == nil {
		return nil
	}
	base := errorBase{
		internal: err,
		message:  fmt.Sprintf(message, formatValues...),
	}
	return BusinessError{
		errorBase: base,
	}
}

type ValidationError struct {
	BusinessError
}

func (ValidationError) Name() string {
	return "ValidationError"
}

func NewValidationError(message string, formatValues ...interface{}) ValidationError {
	base := errorBase{
		errorCode: "ValidationError",
		message:   fmt.Sprintf(message, formatValues...),
	}

	return ValidationError{
		BusinessError{
			errorBase: base,
		},
	}
}

func NewValidationErrorFromOzzo(raw error) ValidationError {
	ozzoErrors, isOk := raw.(apiVal.Errors)
	if !isOk {
		panic(NewTechnicalError("Not an Ozzo validation error"))
	}

	var messages []string
	for _, ozzoErr := range ozzoErrors {
		err := ozzoErr.(apiVal.ErrorObject)
		messages = append(messages, err.Message())
	}

	base := errorBase{
		errorCode: "ValidationError",
		message:   strings.Join(messages, ", "),
	}
	return ValidationError{
		BusinessError{
			errorBase: base,
		},
	}
}

// func WrapValidationError(err error) error {
// 	validationErr, isOk := err.(tagVal.ValidationErrors)
// 	if !isOk {
// 		return nil
// 	}
// 	base := errorBase{
// 		errorCode: errorCode,
// 		internal:  err,
// 		message:   fmt.Sprintf(message, formatValues...),
// 		name:      "BusinessError",
// 	}
// 	return BusinessError{
// 		errorBase: base,
// 	}
// }

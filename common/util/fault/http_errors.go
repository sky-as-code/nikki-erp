package fault

import (
	"fmt"
	"net/http"
)

// HttpError represents an error that occurred while handling a HTTP request.
type HttpError interface {
	error
	// Unwrap supports errors.Unwrap()
	Unwrap() error
	// StatusCode returns HTTP response status code
	StatusCode() int
}

// var ErrorInterface = reflect.TypeOf((*error)(nil)).Elem()

type httpErrorBase struct {
	errorBase
	// The property in response payload to describe the error
	Errors string `json:"errors"`
	status int
}

func (err httpErrorBase) StatusCode() int {
	return err.status
}

// InternalServerHttpError represents an internal server error with status 500
type InternalServerHttpError struct {
	httpErrorBase
}

func NewInternalServerHttpError(message ...interface{}) InternalServerHttpError {
	var msg string = ""
	if len(message) > 0 {
		msg = fmt.Sprint(message[0])
	}

	base := errorBase{
		message: msg,
		name:    "InternalServerHttpError",
	}
	httpErr := InternalServerHttpError{
		httpErrorBase{
			errorBase: base,
			status:    http.StatusInternalServerError,
		},
	}
	httpErr.Errors = httpErr.Error()
	return httpErr
}

func WrapInternalServerHttpError(err error, message ...interface{}) HttpError {
	if err == nil {
		return nil
	}
	var msg string = ""
	if len(message) > 0 {
		msg = fmt.Sprint(message[0])
	}
	base := errorBase{
		internal: err,
		message:  msg,
		name:     "InternalServerHttpError",
	}
	httpErr := InternalServerHttpError{
		httpErrorBase{
			errorBase: base,
			status:    http.StatusInternalServerError,
		},
	}
	httpErr.Errors = httpErr.Error()
	return httpErr
}

type ClientHttpError struct {
	httpErrorBase
}

func NewClientHttpError(message string) ClientHttpError {
	base := errorBase{
		message: message,
		name:    "ClientHttpError",
	}
	httpErr := ClientHttpError{
		httpErrorBase{
			errorBase: base,
			status:    http.StatusUnprocessableEntity,
		},
	}
	httpErr.Errors = httpErr.Error()
	return httpErr
}

func WrapClientHttpError(err error) HttpError {
	if err == nil {
		return nil
	}
	base := errorBase{
		internal: err,
		name:     "ClientHttpError",
	}
	httpErr := ClientHttpError{
		httpErrorBase{
			errorBase: base,
			status:    http.StatusUnprocessableEntity,
		},
	}
	httpErr.Errors = httpErr.Error()
	return httpErr
}

package app

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"go.bryk.io/pkg/errors"
)

func appendClientErrors(dest *ft.ClientErrors, src ft.ClientErrors) {
	for i := range src {
		dest.Append(src[i])
	}
}

func appendNotFoundError(dest *ft.ClientErrors, field string, resourceLabel string) {
	dest.Append(*ft.NewValidationError(field, "", resourceLabel+" not found"))
}

func clientErrorsToError(errs ft.ClientErrors, defaultMessage string) error {
	if errs.Count() == 0 {
		return errors.New(defaultMessage)
	}

	message := errs[0].String()
	if message == "" {
		message = errs[0].Message
	}
	if message == "" {
		message = defaultMessage
	}
	return errors.New(message)
}

package transport

import (
	"errors"

	"github.com/sky-as-code/nikki-erp/modules/core/language/transport/cqrs"
)

func InitTransport() error {
	err := errors.Join(
		cqrs.InitCqrsHandlers(),
	)
	return err
}

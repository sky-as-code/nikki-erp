package transport

import (
	"errors"

	"github.com/sky-as-code/nikki-erp/modules/core/i18n/transport/cqrs"
)

func InitTransport() error {
	err := errors.Join(
		cqrs.InitCqrsHandlers(),
	)
	return err
}

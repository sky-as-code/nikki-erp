package transport

import (
	"errors"

	"github.com/sky-as-code/nikki-erp/modules/drive/transports/restful"
	"github.com/sky-as-code/nikki-erp/modules/drive/transports/cqrs"
)

func InitTransport() error {
	err := errors.Join(
		cqrs.InitCqrsHandlers(),
		restful.InitRestfulHandlers(),
	)

	return err
}

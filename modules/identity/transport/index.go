package transport

import (
	"errors"

	"github.com/sky-as-code/nikki-erp/modules/identity/transport/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/identity/transport/restful"
)

func InitTransport() error {
	err := errors.Join(
		cqrs.InitCqrsHandlers(),
		restful.InitRestfulHandlers(),
	)
	return err
}

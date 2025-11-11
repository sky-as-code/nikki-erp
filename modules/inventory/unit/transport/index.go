package transport

import (
	"errors"

	"github.com/sky-as-code/nikki-erp/modules/inventory/unit/transport/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/inventory/unit/transport/restful"
)

func InitTransport() error {
	err := errors.Join(
		cqrs.InitCqrsHandlers(),
		restful.InitRestfulHandlers(),
	)
	return err
}

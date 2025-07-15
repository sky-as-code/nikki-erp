package transport

import (
	"errors"

	"github.com/sky-as-code/nikki-erp/modules/contacts/transport/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/contacts/transport/restful"
)

func InitTransport() error {
	err := errors.Join(
		cqrs.InitCqrsHandlers(),
		restful.InitRestfulHandlers(),
	)
	return err
}

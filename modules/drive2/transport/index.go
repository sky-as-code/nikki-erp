package transport

import (
	"errors"

	"github.com/sky-as-code/nikki-erp/modules/drive2/transport/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/drive2/transport/restful"
)

func InitTransport() error {
	return errors.Join(
		cqrs.InitCqrsHandlers(),
		restful.InitRestfulHandlers(),
	)
}

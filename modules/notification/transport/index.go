package transport

import (
	stdErr "errors"

	"github.com/sky-as-code/nikki-erp/modules/notification/transport/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/notification/transport/restful"
)

func InitTransport() error {
	return stdErr.Join(
		cqrs.InitCqrsHandlers(),
		restful.InitRestfulHandlers(),
	)
}

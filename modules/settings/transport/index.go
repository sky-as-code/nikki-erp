package transport

import (
	stdErr "errors"

	"github.com/sky-as-code/nikki-erp/modules/settings/transport/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/settings/transport/restful"
)

func InitTransport() error {
	return stdErr.Join(
		cqrs.InitCqrsHandlers(),
		restful.InitRestfulHandlers(),
	)
}

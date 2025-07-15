package transport

import (
	"errors"

	"github.com/sky-as-code/nikki-erp/modules/authenticate/transport/restful"
)

func InitTransport() error {
	err := errors.Join(
		restful.InitRestfulHandlers(),
	)
	return err
}

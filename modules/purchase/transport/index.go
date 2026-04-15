package transport

import (
	"errors"

	"github.com/sky-as-code/nikki-erp/modules/inventory/product/transport/restful"
)

func InitTransport() error {
	err := errors.Join(
		restful.InitRestfulHandlers(),
	)
	return err
}

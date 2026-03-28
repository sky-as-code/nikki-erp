package adapter

import (
	"errors"

	"github.com/sky-as-code/nikki-erp/modules/drive/adapter/cqrs_bus"
	"github.com/sky-as-code/nikki-erp/modules/drive/adapter/external"
)

func InitAdapters() error {
	err := errors.Join(
		cqrs_bus.InitCqrsBusAdaper(),
		external.InitExternalAdapter(),
	)

	return err
}

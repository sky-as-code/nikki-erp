package adapter

import (
	"errors"

	"github.com/sky-as-code/nikki-erp/modules/drive2/adapter/cqrs_bus"
	"github.com/sky-as-code/nikki-erp/modules/drive2/infra/external"
)

func InitAdapters() error {
	return errors.Join(
		cqrs_bus.InitCqrsBusAdapter(),
		external.InitExternalAdapter(),
	)
}

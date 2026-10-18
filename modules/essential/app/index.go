package app

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	itUom "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uom"
)

func InitApplicationServices() error {
	return stdErr.Join(
		deps.Register(
			NewSettingsApplicationServiceImpl,
			NewUomConversionApplicationServiceImpl,
		),
		// The narrow port sales, purchase, inventory and accounting bind to.
		deps.Register(func(svc itUom.UomConversionApplicationService) itUom.UomConversionAppService {
			return svc
		}),
	)
}

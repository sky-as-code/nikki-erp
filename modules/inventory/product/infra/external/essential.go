package external

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	itUnit "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/unit"
	itExt "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/external"
)

func InitExternal() error {
	err := stdErr.Join(
		deps.Register(func(unitSvc itUnit.UnitService) itExt.UnitExtService {
			// This will be replaced with the actual implementation when this application is
			// split into separate microservices.
			return unitSvc
		}),
	)

	return err
}

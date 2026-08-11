package external

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	itUom "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uom"
	itExt "github.com/sky-as-code/nikki-erp/modules/inventory/legacy/interfaces/external"
)

// This file is the only place in the product feature where an import of another module may
// appear. Everything else depends on the local port in interfaces/external.

func InitExternal() error {
	err := stdErr.Join(
		deps.Register(func(uomSvc itUom.UomConversionAppService) itExt.UomExtService {
			// This will be replaced with a REST/CQRS client when this application is
			// split into separate microservices.
			return uomSvc
		}),
	)

	return err
}
